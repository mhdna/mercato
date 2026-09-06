package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mhdna/kashi/db/sqlc"
)

// Employees are a company-wide HR roster with a single current payroll
// snapshot each (see db/migrations/000074_create_employees.up.sql). The
// deduction (rate * late units) and the net salary (base + commission -
// deduction) are pure arithmetic recomputed here on every read, never
// stored -- employeeResponse is the one place that math lives.

const (
	lateUnitsCentiPerUnit = 100
	dateLayout            = "2006-01-02"
)

// employeeResponse is the wire shape for an employee: the stored columns
// with sql.Null* unwrapped to plain JSON, plus the two computed money
// fields the payroll table renders and colours.
type employeeResponse struct {
	ID                     int64     `json:"id"`
	BranchID               *int64    `json:"branch_id"`
	Name                   string    `json:"name"`
	Role                   string    `json:"role"`
	Status                 string    `json:"status"`
	HiredOn                *string   `json:"hired_on"`
	BaseSalaryCents        int64     `json:"base_salary_cents"`
	CommissionCents        int64     `json:"commission_cents"`
	LateDeductionRateCents int64     `json:"late_deduction_rate_cents"`
	LateDeductionUnit      string    `json:"late_deduction_unit"`
	LateUnitsCenti         int64     `json:"late_units_centi"`
	DeductionCents         int64     `json:"deduction_cents"`
	NetSalaryCents         int64     `json:"net_salary_cents"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func newEmployeeResponse(e db.Employee) employeeResponse {
	deduction := e.LateDeductionRateCents * e.LateUnitsCenti / lateUnitsCentiPerUnit

	res := employeeResponse{
		ID:                     e.ID,
		Name:                   e.Name,
		Role:                   e.Role,
		Status:                 e.Status,
		BaseSalaryCents:        e.BaseSalaryCents,
		CommissionCents:        e.CommissionCents,
		LateDeductionRateCents: e.LateDeductionRateCents,
		LateDeductionUnit:      e.LateDeductionUnit,
		LateUnitsCenti:         e.LateUnitsCenti,
		DeductionCents:         deduction,
		NetSalaryCents:         e.BaseSalaryCents + e.CommissionCents - deduction,
		CreatedAt:              e.CreatedAt,
		UpdatedAt:              e.UpdatedAt,
	}
	if e.BranchID.Valid {
		res.BranchID = &e.BranchID.Int64
	}
	if e.HiredOn.Valid {
		s := e.HiredOn.Time.Format(dateLayout)
		res.HiredOn = &s
	}
	return res
}

func newEmployeeResponses(rows []db.Employee) []employeeResponse {
	out := make([]employeeResponse, len(rows))
	for i, e := range rows {
		out[i] = newEmployeeResponse(e)
	}
	return out
}

// nullableBranchID maps the optional branch_id from a request (0 = none)
// to the nullable column value.
func nullableBranchID(id int64) sql.NullInt64 {
	if id > 0 {
		return sql.NullInt64{Int64: id, Valid: true}
	}
	return sql.NullInt64{}
}

// parseHiredOn accepts "" (unset) or a YYYY-MM-DD date.
func parseHiredOn(s string) (sql.NullTime, error) {
	if s == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return sql.NullTime{}, err
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

type employeeBody struct {
	BranchID               int64  `json:"branch_id"`
	Name                   string `json:"name" binding:"required"`
	Role                   string `json:"role"`
	Status                 string `json:"status" binding:"omitempty,oneof=active inactive"`
	HiredOn                string `json:"hired_on"`
	BaseSalaryCents        int64  `json:"base_salary_cents" binding:"min=0"`
	CommissionCents        int64  `json:"commission_cents" binding:"min=0"`
	LateDeductionRateCents int64  `json:"late_deduction_rate_cents" binding:"min=0"`
	LateDeductionUnit      string `json:"late_deduction_unit" binding:"omitempty,oneof=hour day"`
	LateUnitsCenti         int64  `json:"late_units_centi" binding:"min=0"`
}

func (b employeeBody) status() string {
	if b.Status == "" {
		return "active"
	}
	return b.Status
}

func (b employeeBody) lateUnit() string {
	if b.LateDeductionUnit == "" {
		return "hour"
	}
	return b.LateDeductionUnit
}

func (server *Server) createEmployee(ctx *gin.Context) {
	var req employeeBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	hiredOn, err := parseHiredOn(req.HiredOn)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	employee, err := server.store.CreateEmployee(ctx, db.CreateEmployeeParams{
		BranchID:               nullableBranchID(req.BranchID),
		Name:                   req.Name,
		Role:                   req.Role,
		Status:                 req.status(),
		HiredOn:                hiredOn,
		BaseSalaryCents:        req.BaseSalaryCents,
		CommissionCents:        req.CommissionCents,
		LateDeductionRateCents: req.LateDeductionRateCents,
		LateDeductionUnit:      req.lateUnit(),
		LateUnitsCenti:         req.LateUnitsCenti,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"employee": newEmployeeResponse(employee)})
}

type employeeIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getEmployee(ctx *gin.Context) {
	var uri employeeIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	employee, err := server.store.GetEmployee(ctx, uri.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"employee": newEmployeeResponse(employee)})
}

func (server *Server) updateEmployee(ctx *gin.Context) {
	var uri employeeIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	var req employeeBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	hiredOn, err := parseHiredOn(req.HiredOn)
	if err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	employee, err := server.store.UpdateEmployee(ctx, db.UpdateEmployeeParams{
		ID:                     uri.ID,
		BranchID:               nullableBranchID(req.BranchID),
		Name:                   req.Name,
		Role:                   req.Role,
		Status:                 req.status(),
		HiredOn:                hiredOn,
		BaseSalaryCents:        req.BaseSalaryCents,
		CommissionCents:        req.CommissionCents,
		LateDeductionRateCents: req.LateDeductionRateCents,
		LateDeductionUnit:      req.lateUnit(),
		LateUnitsCenti:         req.LateUnitsCenti,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"employee": newEmployeeResponse(employee)})
}

func (server *Server) deleteEmployee(ctx *gin.Context) {
	var uri employeeIDRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	if _, err := server.store.GetEmployee(ctx, uri.ID); err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}

	if err := server.store.DeleteEmployee(ctx, uri.ID); err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"message": "employee deleted"})
}

type listEmployeesRequest struct {
	listPageQuery
	BranchID int64 `form:"branch_id"`
}

func (server *Server) listEmployees(ctx *gin.Context) {
	var req listEmployeesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}
	branchID := nullableBranchID(req.BranchID)

	rows, err := server.store.ListEmployees(ctx, db.ListEmployeesParams{
		Search:     req.Search,
		BranchID:   branchID,
		PageSize:   req.PageSize,
		PageOffset: req.PageID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	total, err := server.store.CountEmployees(ctx, db.CountEmployeesParams{
		Search:   req.Search,
		BranchID: branchID,
	})
	if err != nil {
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"employees": newEmployeeResponses(rows), "total": total})
}

type batchSalaryItem struct {
	ID                     int64  `json:"id" binding:"required,min=1"`
	BaseSalaryCents        int64  `json:"base_salary_cents" binding:"min=0"`
	CommissionCents        int64  `json:"commission_cents" binding:"min=0"`
	LateDeductionRateCents int64  `json:"late_deduction_rate_cents" binding:"min=0"`
	LateDeductionUnit      string `json:"late_deduction_unit" binding:"required,oneof=hour day"`
	LateUnitsCenti         int64  `json:"late_units_centi" binding:"min=0"`
}

type batchSalaryRequest struct {
	Items []batchSalaryItem `json:"items" binding:"required,min=1,dive"`
}

// batchUpdateEmployeeSalaries is the "Save all" action of the payroll
// table's edit mode -- every changed row in one atomic write.
func (server *Server) batchUpdateEmployeeSalaries(ctx *gin.Context) {
	var req batchSalaryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.writeError(ctx, http.StatusBadRequest, err)
		return
	}

	args := make([]db.UpdateEmployeeSalaryParams, len(req.Items))
	for i, item := range req.Items {
		args[i] = db.UpdateEmployeeSalaryParams{
			ID:                     item.ID,
			BaseSalaryCents:        item.BaseSalaryCents,
			CommissionCents:        item.CommissionCents,
			LateDeductionRateCents: item.LateDeductionRateCents,
			LateDeductionUnit:      item.LateDeductionUnit,
			LateUnitsCenti:         item.LateUnitsCenti,
		}
	}

	updated, err := server.store.BatchSetEmployeeSalaries(ctx, args)
	if err != nil {
		if err == sql.ErrNoRows {
			server.writeError(ctx, http.StatusNotFound, err)
			return
		}
		server.writeError(ctx, http.StatusInternalServerError, err)
		return
	}
	server.writeJSON(ctx, http.StatusOK, envelope{"employees": newEmployeeResponses(updated)})
}
