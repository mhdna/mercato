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
	// ColorIDs/SizeIDs let one product be created with several variants at
	// once -- one per colour x size combination. When both are empty the
	// single ColorID/SizeID (or neither) is used, keeping the old
	// one-variant behaviour.
	ColorIDs []int64 `json:"color_ids"`
	SizeIDs  []int64 `json:"size_ids"`
	Barcode  *int64  `json:"barcode"`
}

type CreateProductTxResult struct {
	Product           Product             `json:"product"`
	ProductAttributes []ProductsAttribute `json:"product_attributes"`
	// Variant is the first created variant, kept for existing callers.
	Variant  ProductVariant   `json:"variant"`
	Variants []ProductVariant `json:"variants"`
}

// createVariantWithBarcode inserts one variant, auto-generating a unique
// EAN-13 barcode (retrying on collision) unless an explicit barcode is given.
func createVariantWithBarcode(ctx context.Context, q *Queries, productID int64, colorID, sizeID, price sql.NullInt64, explicitBarcode *int64) (ProductVariant, error) {
	if explicitBarcode != nil {
		return q.CreateProductVariant(ctx, CreateProductVariantParams{
			ProductID: productID,
			ColorID:   colorID,
			SizeID:    sizeID,
			Barcode:   strconv.FormatInt(*explicitBarcode, 10),
			Price:     price,
		})
	}

	for attempt := 0; attempt < 20; attempt++ {
		itemNumber, err := q.GetNextBarcodeItemValue(ctx)
		if err != nil {
			return ProductVariant{}, err
		}
		variant, err := q.CreateProductVariant(ctx, CreateProductVariantParams{
			ProductID: productID,
			ColorID:   colorID,
			SizeID:    sizeID,
			Barcode:   strconv.FormatInt(generateEAN13(itemNumber), 10),
			Price:     price,
		})
		if err == nil {
			return variant, nil
		}
		var pqErr *pq.Error
		if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
			return ProductVariant{}, err
		}
	}
	return ProductVariant{}, fmt.Errorf("failed to generate unique barcode after 20 attempts")
}

func nullInt64(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

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

		price := sql.NullInt64{}
		if arg.Price != 0 {
			price = sql.NullInt64{Int64: arg.Price, Valid: true}
		}

		// Build the set of colour ids and size ids to fan out over. An
		// empty slice means "no dimension" -> a single nil entry.
		colorIDs := arg.ColorIDs
		if len(colorIDs) == 0 && arg.ColorID != nil {
			colorIDs = []int64{*arg.ColorID}
		}
		sizeIDs := arg.SizeIDs
		if len(sizeIDs) == 0 && arg.SizeID != nil {
			sizeIDs = []int64{*arg.SizeID}
		}
		colorCells := []sql.NullInt64{{}}
		if len(colorIDs) > 0 {
			colorCells = colorCells[:0]
			for _, id := range colorIDs {
				colorCells = append(colorCells, nullInt64(id))
			}
		}
		sizeCells := []sql.NullInt64{{}}
		if len(sizeIDs) > 0 {
			sizeCells = sizeCells[:0]
			for _, id := range sizeIDs {
				sizeCells = append(sizeCells, nullInt64(id))
			}
		}

		// An explicit barcode can only apply to a single variant.
		explicitBarcode := arg.Barcode
		if explicitBarcode != nil && len(colorCells)*len(sizeCells) > 1 {
			return fmt.Errorf("an explicit barcode cannot be used when creating multiple variants")
		}

		for _, c := range colorCells {
			for _, s := range sizeCells {
				variant, err := createVariantWithBarcode(ctx, q, product.ID, c, s, price, explicitBarcode)
				if err != nil {
					return err
				}
				result.Variants = append(result.Variants, variant)
			}
		}

		result.Product = product
		result.ProductAttributes = productAttributes
		result.Variant = result.Variants[0]

		return nil
	})

	return result, err
}

type UpdateProductTxParams struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	// AttributeValues carries the attribute id (0-8) plus its free-text
	// value; a blank value clears nothing (it's simply skipped) so callers
	// only need to send the attributes they want set.
	AttributeValues []AttributesValue `json:"attribute_values"`
}

type UpdateProductTxResult struct {
	Product           Product             `json:"product"`
	ProductAttributes []ProductsAttribute `json:"product_attributes"`
}

// UpdateProductTx edits the shared product fields and upserts its
// attributes in one transaction -- the write-side mirror of
// CreateProductTx's attribute handling, for products that already exist.
func (store *SQLStore) UpdateProductTx(ctx context.Context, arg UpdateProductTxParams) (UpdateProductTxResult, error) {
	var result UpdateProductTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		if err := q.UpdateProduct(ctx, UpdateProductParams{
			ID:          arg.ID,
			Code:        arg.Code,
			Name:        arg.Name,
			Description: arg.Description,
			IsActive:    arg.IsActive,
		}); err != nil {
			return err
		}

		for _, a := range arg.AttributeValues {
			if a.Value == "" {
				continue
			}
			attributeValue, err := q.UpsertAttributeValue(ctx, UpsertAttributeValueParams{
				AttributeID: a.AttributeID,
				Value:       a.Value,
			})
			if err != nil {
				return err
			}
			if err := q.UpsertProductAttribute(ctx, UpsertProductAttributeParams{
				ProductID:        arg.ID,
				AttributeID:      a.AttributeID,
				AttributeValueID: attributeValue.ID,
			}); err != nil {
				return err
			}
		}

		product, err := q.GetProduct(ctx, arg.ID)
		if err != nil {
			return err
		}
		productAttributes, err := q.GetProductAttributes(ctx, arg.ID)
		if err != nil {
			return err
		}
		result.Product = product
		result.ProductAttributes = productAttributes
		return nil
	})

	return result, err
}

type AddVariantTxParams struct {
	ProductID int64  `json:"product_id"`
	ColorID   *int64 `json:"color_id"`
	SizeID    *int64 `json:"size_id"`
	Price     *int64 `json:"price"`
	// Barcode is optional -- a unique EAN-13 is generated when it's nil.
	Barcode *int64 `json:"barcode"`
}

// AddVariantTx adds a single variant to an existing product, reusing the
// same auto-barcode generation CreateProductTx uses for the first variant.
func (store *SQLStore) AddVariantTx(ctx context.Context, arg AddVariantTxParams) (ProductVariant, error) {
	var variant ProductVariant

	err := store.execTx(ctx, func(q *Queries) error {
		colorID := sql.NullInt64{}
		if arg.ColorID != nil {
			colorID = nullInt64(*arg.ColorID)
		}
		sizeID := sql.NullInt64{}
		if arg.SizeID != nil {
			sizeID = nullInt64(*arg.SizeID)
		}
		price := sql.NullInt64{}
		if arg.Price != nil {
			price = nullInt64(*arg.Price)
		}

		v, err := createVariantWithBarcode(ctx, q, arg.ProductID, colorID, sizeID, price, arg.Barcode)
		if err != nil {
			return err
		}
		variant = v
		return nil
	})

	return variant, err
}
