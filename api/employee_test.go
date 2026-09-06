package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mhdna/kashi/db/mock"
	db "github.com/mhdna/kashi/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestCreateEmployee_OK_DefaultsAndNet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().
		CreateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, arg db.CreateEmployeeParams) (db.Employee, error) {
			require.Equal(t, "active", arg.Status)          // defaulted
			require.Equal(t, "hour", arg.LateDeductionUnit) // defaulted
			require.False(t, arg.BranchID.Valid)            // 0 -> NULL
			return db.Employee{
				ID: 1, Name: arg.Name, Role: arg.Role, Status: arg.Status,
				BaseSalaryCents: arg.BaseSalaryCents, CommissionCents: arg.CommissionCents,
				LateDeductionRateCents: arg.LateDeductionRateCents,
				LateDeductionUnit:      arg.LateDeductionUnit,
				LateUnitsCenti:         arg.LateUnitsCenti,
			}, nil
		})

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPost, "/employees", map[string]any{
		"name":                      "Sara",
		"role":                      "Salesman",
		"base_salary_cents":         100000, // $1000.00
		"commission_cents":          25000,  // $250.00
		"late_deduction_rate_cents": 200,    // $2.00 / hour
		"late_units_centi":          250,    // 2.5 hours late
	})

	require.Equal(t, http.StatusOK, rec.Code)
	var got struct {
		Employee employeeResponse `json:"employee"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, int64(500), got.Employee.DeductionCents)    // 200 * 250 / 100
	require.Equal(t, int64(124500), got.Employee.NetSalaryCents) // 100000 + 25000 - 500
	require.Nil(t, got.Employee.BranchID)
}

func TestCreateEmployee_RejectsBadStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().CreateEmployee(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPost, "/employees", map[string]any{
		"name": "Sara", "status": "retired",
	})
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListEmployees_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().
		ListEmployees(gomock.Any(), gomock.Any()).
		Return([]db.Employee{
			{ID: 1, Name: "Sara", BaseSalaryCents: 100000, CommissionCents: 0, LateDeductionUnit: "hour"},
		}, nil)
	store.EXPECT().CountEmployees(gomock.Any(), gomock.Any()).Return(int64(1), nil)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodGet, "/employees?page_size=10", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	var got struct {
		Employees []employeeResponse `json:"employees"`
		Total     int64              `json:"total"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, int64(1), got.Total)
	require.Len(t, got.Employees, 1)
	require.Equal(t, int64(100000), got.Employees[0].NetSalaryCents)
}

func TestBatchUpdateEmployeeSalaries_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)

	store.EXPECT().
		BatchSetEmployeeSalaries(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, args []db.UpdateEmployeeSalaryParams) ([]db.Employee, error) {
			require.Len(t, args, 2)
			out := make([]db.Employee, len(args))
			for i, a := range args {
				out[i] = db.Employee{
					ID: a.ID, BaseSalaryCents: a.BaseSalaryCents, CommissionCents: a.CommissionCents,
					LateDeductionRateCents: a.LateDeductionRateCents, LateDeductionUnit: a.LateDeductionUnit,
					LateUnitsCenti: a.LateUnitsCenti,
				}
			}
			return out, nil
		})

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPut, "/employees/salaries", map[string]any{
		"items": []map[string]any{
			{"id": 1, "base_salary_cents": 100000, "commission_cents": 0, "late_deduction_rate_cents": 0, "late_deduction_unit": "hour", "late_units_centi": 0},
			{"id": 2, "base_salary_cents": 90000, "commission_cents": 10000, "late_deduction_rate_cents": 500, "late_deduction_unit": "day", "late_units_centi": 100},
		},
	})

	require.Equal(t, http.StatusOK, rec.Code)
	var got struct {
		Employees []employeeResponse `json:"employees"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Len(t, got.Employees, 2)
	require.Equal(t, int64(99500), got.Employees[1].NetSalaryCents) // 90000 + 10000 - (500*100/100)
}

func TestBatchUpdateEmployeeSalaries_RejectsBadUnit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mockdb.NewMockStore(ctrl)
	store.EXPECT().BatchSetEmployeeSalaries(gomock.Any(), gomock.Any()).Times(0)

	server := newTestServer(t, store)
	rec := adminRequest(t, server, http.MethodPut, "/employees/salaries", map[string]any{
		"items": []map[string]any{
			{"id": 1, "base_salary_cents": 100000, "late_deduction_unit": "week"},
		},
	})
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
