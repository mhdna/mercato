package db

import (
	"context"
	"database/sql"
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
	Variant           ProductVariant      `json:"variant"`
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

// CreateProductTx creates a product together with its first variant. A
// product is never created without at least one sellable variant — a
// product with zero variants has no barcode a till could ring up, so
// there'd be nothing to actually sell.
func (store *SQLStore) CreateProductTx(ctx context.Context, arg CreateProductTxParams) (CreateProductTxResult, error) {
	var result CreateProductTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		product, err := q.CreateProduct(ctx, CreateProductParams{
			Code:        arg.Code,
			Name:        arg.Name,
			Description: arg.Description,
		})
		if err != nil {
			return err
		}

		productAttributes := []ProductsAttribute{}
		for _, a := range arg.AttributeValues {
			attributeValue, err := q.UpsertAttributeValue(ctx, UpsertAttributeValueParams{
				AttributeID: a.AttributeID,
				Value:       a.Value,
			})
			if err != nil {
				return err
			}

			productAttribute, err := q.CreateProductAttribute(ctx, CreateProductAttributeParams{
				ProductID:        product.ID,
				AttributeID:      a.AttributeID,
				AttributeValueID: attributeValue.ID,
			})
			if err != nil {
				return err
			}
			productAttributes = append(productAttributes, productAttribute)
		}

		colorID := sql.NullInt64{}
		if arg.ColorID != nil {
			colorID = sql.NullInt64{Int64: *arg.ColorID, Valid: true}
		}
		sizeID := sql.NullInt64{}
		if arg.SizeID != nil {
			sizeID = sql.NullInt64{Int64: *arg.SizeID, Valid: true}
		}
		price := sql.NullInt64{}
		if arg.Price != 0 {
			price = sql.NullInt64{Int64: arg.Price, Valid: true}
		}

		var variant ProductVariant

		if arg.Barcode != nil {
			variant, err = q.CreateProductVariant(ctx, CreateProductVariantParams{
				ProductID: product.ID,
				ColorID:   colorID,
				SizeID:    sizeID,
				Barcode:   strconv.FormatInt(*arg.Barcode, 10),
				Price:     price,
			})
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
				variant, err = q.CreateProductVariant(ctx, CreateProductVariantParams{
					ProductID: product.ID,
					ColorID:   colorID,
					SizeID:    sizeID,
					Barcode:   strconv.FormatInt(barcodeValue, 10),
					Price:     price,
				})
				if err == nil {
					inserted = true
					break
				}

				var pqErr *pq.Error
				if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
					return err
				}
			}
			if !inserted {
				return fmt.Errorf("failed to generate unique barcode after 20 attempts")
			}
		}

		result.Product = product
		result.ProductAttributes = productAttributes
		result.Variant = variant

		return nil
	})

	return result, err
}
