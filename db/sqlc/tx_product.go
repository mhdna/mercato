package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/lib/pq"
)

type CreateProductTxParams struct {
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Price           int64             `json:"price"`
	Discount        int16             `json:"discount"`
	AttributeValues []AttributesValue `json:"attribute_values"`
	ColorID         *int64            `json:"color_id"`
	SizeID          *int64            `json:"size_id"`
	Barcode         *int64            `json:"barcode"`
}

type CreateProductTxResult struct {
	Product           Product             `json:"product"`
	ProductAttributes []ProductsAttribute `json:"product_attributes"`
	Barcode           Barcode             `json:"barcode"`
	ProductColor      *ProductsColor      `json:"product_color,omitempty"`
	ProductSize       *ProductsSize       `json:"product_size,omitempty"`
}

func generateEAN13(itemNumber int64) int64 {
	base := fmt.Sprintf("800%09d", itemNumber)
	sum := 0
	for i, ch := range base {
		d := int(ch - '0')
		if i%2 == 0 {
			sum += d * 3
		} else {
			sum += d
		}
	}
	check := (10 - sum%10) % 10
	result, _ := strconv.ParseInt(base+strconv.Itoa(check), 10, 64)
	return result
}

func (store *SQLStore) CreateProductTx(ctx context.Context, arg CreateProductTxParams) (CreateProductTxResult, error) {
	var result CreateProductTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		createProductArg := CreateProductParams{
			Code:        arg.Code,
			Name:        arg.Name,
			Description: arg.Description,
		}

		product, err := q.CreateProduct(ctx, createProductArg)
		if err != nil {
			return err
		}

		productAttributes := []ProductsAttribute{}

		for _, a := range arg.AttributeValues {
			// make sure attribute value exists
			upsertAttributeValueArg := UpsertAttributeValueParams{
				AttributeID: a.AttributeID,
				Value:       a.Value,
			}
			attributeValue, err := q.UpsertAttributeValue(ctx, upsertAttributeValueArg)
			if err != nil {
				return err
			}

			// then add attribute value to product
			addAttributeArg := CreateProductAttributeParams{
				ProductID:        product.ID,
				AttributeID:      a.AttributeID,
				AttributeValueID: attributeValue.ID,
			}
			fmt.Println(a.AttributeID)
			fmt.Println(a.ID)
			fmt.Println(product.ID)
			productAttribute, err := q.CreateProductAttribute(ctx, addAttributeArg)

			if err != nil {
				return err
			}
			productAttributes = append(productAttributes, productAttribute)
		}

		var barcode Barcode

		if arg.Barcode != nil {
			barcodeParams := CreateBarcodeParams{
				ProductID: product.ID,
				Barcode:   *arg.Barcode,
				// ColorID:   sql.NullInt64{Valid: true, Int64: *arg.ColorID},
				// SizeID:    sql.NullInt64{Valid: true, Int64: *arg.SizeID},
			}
			barcode, err = q.CreateBarcode(ctx, barcodeParams)
			if err != nil {
				return err
			}
		} else {
			inserted := false
			for attempt := 0; attempt < 20; attempt++ {
				itemNumber, err := q.GetNextBarcodeItemValue(ctx)
				if err != nil {
					return err
				}

				barcodeValue := generateEAN13(itemNumber)
				barcode, err = q.CreateBarcode(ctx, CreateBarcodeParams{
					Barcode: barcodeValue, ProductID: product.ID,
				})
				if err == nil {
					inserted = true
					break
				}

				// TODO: centralize error handling
				var pqErr *pq.Error
				if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
					return err
				}
			}
			if !inserted {
				return fmt.Errorf("failed to generate unique barcode after 20 attempts")
			}
		}

		if arg.ColorID != nil {
			productColor, err := q.CreateProductColor(ctx, CreateProductColorParams{
				ProductID: product.ID,
				ColorID:   *arg.ColorID,
			})
			if err != nil {
				return err
			}
			result.ProductColor = &productColor
		}

		if arg.SizeID != nil {
			productSize, err := q.CreateProductSize(ctx, CreateProductSizeParams{
				ProductID: product.ID,
				SizeID:    *arg.SizeID,
			})
			if err != nil {
				return err
			}
			result.ProductSize = &productSize
		}

		result.Product = product
		result.ProductAttributes = productAttributes
		result.Barcode = barcode

		return nil
	})

	return result, err
}
