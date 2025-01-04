package repositorys

import (
	"context"
	"database/sql"
	"discountdb-api/internal/models"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"
)

type CouponRepository struct {
	db *sql.DB
}

func NewCouponRepository(db *sql.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// Migration SQL to create the table
const createTableSQL = `
CREATE TABLE IF NOT EXISTS coupons (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    code VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount_value DECIMAL(10,2) NOT NULL,
    discount_type VARCHAR(50) NOT NULL,
    merchant_name VARCHAR(255) NOT NULL,
    merchant_url TEXT NOT NULL,
    
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    terms_conditions TEXT,
    minimum_purchase_amount DECIMAL(10,2),
    maximum_discount_amount DECIMAL(10,2),
    
    up_votes TIMESTAMP[] DEFAULT ARRAY[]::TIMESTAMP[],
    down_votes TIMESTAMP[] DEFAULT ARRAY[]::TIMESTAMP[],
    
    categories TEXT[] DEFAULT ARRAY[]::TEXT[],
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    regions TEXT[] DEFAULT ARRAY[]::TEXT[],
    store_type VARCHAR(50),
    
    materialized_score DECIMAL(10,4),
    last_score_update TIMESTAMP,
    
    CONSTRAINT valid_discount_type CHECK (
        discount_type IN ('PERCENTAGE_OFF', 'FIXED_AMOUNT', 'BOGO', 'FREE_SHIPPING')
    ),
    CONSTRAINT valid_store_type CHECK (
        store_type IN ('online', 'in_store', 'both')
    )
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_coupons_merchant ON coupons(merchant_name);
CREATE INDEX IF NOT EXISTS idx_coupons_created_at ON coupons(created_at);
CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_score ON coupons(materialized_score);

CREATE OR REPLACE FUNCTION calculate_coupon_score(
    p_discount_value DECIMAL,
    p_discount_type VARCHAR,
    p_maximum_discount_amount DECIMAL,
    p_created_at TIMESTAMP,
    p_up_votes TIMESTAMP[],
    p_down_votes TIMESTAMP[]
) RETURNS DECIMAL AS $$
DECLARE
    vote_score DECIMAL;
    discount_score DECIMAL;
    freshness_score DECIMAL;
BEGIN
    -- Calculate vote score
    SELECT COALESCE(
        SUM(
            CASE 
                WHEN age < INTERVAL '1 day' THEN 1.0
                WHEN age < INTERVAL '1 week' THEN 0.8
                WHEN age < INTERVAL '1 month' THEN 0.6
                WHEN age < INTERVAL '6 months' THEN 0.4
                ELSE 0.2
            END
        ), 0) INTO vote_score
    FROM (
        SELECT CURRENT_TIMESTAMP - unnest(p_up_votes) as age
    ) up;
    
    SELECT vote_score - COALESCE(
        SUM(
            CASE 
                WHEN age < INTERVAL '1 day' THEN 1.0
                WHEN age < INTERVAL '1 week' THEN 0.8
                WHEN age < INTERVAL '1 month' THEN 0.6
                WHEN age < INTERVAL '6 months' THEN 0.4
                ELSE 0.2
            END
        ), 0) INTO vote_score
    FROM (
        SELECT CURRENT_TIMESTAMP - unnest(p_down_votes) as age
    ) down;

    -- Calculate discount score
    discount_score := CASE 
        WHEN p_discount_type = 'PERCENTAGE_OFF' THEN 
            LEAST(p_discount_value / 100.0, 1.0)
        WHEN p_discount_type = 'FIXED_AMOUNT' THEN 
            CASE 
                WHEN p_maximum_discount_amount > 0 THEN 
                    LEAST(p_discount_value / p_maximum_discount_amount, 1.0)
                ELSE 
                    LEAST(p_discount_value / 1000.0, 1.0)
            END
        WHEN p_discount_type IN ('BOGO', 'FREE_SHIPPING') THEN 
            0.5
    END;

    -- Calculate freshness score
    freshness_score := CASE 
        WHEN CURRENT_TIMESTAMP - p_created_at < INTERVAL '1 day' THEN 1.0
        WHEN CURRENT_TIMESTAMP - p_created_at < INTERVAL '1 week' THEN 0.8
        WHEN CURRENT_TIMESTAMP - p_created_at < INTERVAL '1 month' THEN 0.6
        WHEN CURRENT_TIMESTAMP - p_created_at < INTERVAL '3 months' THEN 0.4
        WHEN CURRENT_TIMESTAMP - p_created_at < INTERVAL '6 months' THEN 0.2
        ELSE 0.1
    END;

    -- Return weighted score
    RETURN (vote_score * 0.4) + (discount_score * 0.4) + (freshness_score * 0.2);
END;
$$ LANGUAGE plpgsql;

-- Create batch update function
CREATE OR REPLACE FUNCTION update_materialized_scores_batch(batch_size INT) 
RETURNS void AS $$
BEGIN
    WITH coupons_to_update AS (
        SELECT id, discount_value, discount_type, maximum_discount_amount, 
               created_at, up_votes, down_votes
        FROM coupons 
        WHERE last_score_update IS NULL
        OR last_score_update < CURRENT_TIMESTAMP - INTERVAL '1 hour'
        ORDER BY last_score_update NULLS FIRST 
        LIMIT batch_size
        FOR UPDATE SKIP LOCKED
    )
    UPDATE coupons c
    SET materialized_score = calculate_coupon_score(
            ct.discount_value,
            ct.discount_type,
            ct.maximum_discount_amount,
            ct.created_at,
            ct.up_votes,
            ct.down_votes
        ),
        last_score_update = CURRENT_TIMESTAMP
    FROM coupons_to_update ct
    WHERE c.id = ct.id;
END;
$$ LANGUAGE plpgsql;

-- Create trigger function
CREATE OR REPLACE FUNCTION update_coupon_score() RETURNS TRIGGER AS $$
BEGIN
    NEW.materialized_score := calculate_coupon_score(
        NEW.discount_value,
        NEW.discount_type,
        NEW.maximum_discount_amount,
        NEW.created_at,
        NEW.up_votes,
        NEW.down_votes
    );
    NEW.last_score_update := CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS update_score_trigger ON coupons;
CREATE TRIGGER update_score_trigger
    BEFORE INSERT OR UPDATE OF discount_value, discount_type, maximum_discount_amount, up_votes, down_votes
    ON coupons
    FOR EACH ROW
    EXECUTE FUNCTION update_coupon_score();
`

func (r *CouponRepository) CreateTable(ctx context.Context) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Execute the entire script
	_, err = tx.ExecContext(ctx, createTableSQL)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create tables and functions: %v", err)
	}

	return tx.Commit()
}

func (r *CouponRepository) Create(ctx context.Context, coupon *models.Coupon) error {
	const query = `
        INSERT INTO coupons (
            code, title, description, discount_value, discount_type,
            merchant_name, merchant_url, start_date, end_date,
            terms_conditions, minimum_purchase_amount, maximum_discount_amount,
            up_votes, down_votes, categories, tags, regions, store_type
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
            $13, $14, $15, $16, $17, $18
        ) RETURNING id, created_at, materialized_score`

	return r.db.QueryRowContext(ctx, query,
		coupon.Code, coupon.Title, coupon.Description,
		coupon.DiscountValue, coupon.DiscountType,
		coupon.MerchantName, coupon.MerchantURL,
		coupon.StartDate, coupon.EndDate,
		coupon.TermsConditions, coupon.MinimumPurchaseAmount,
		coupon.MaximumDiscountAmount, &coupon.UpVotes,
		&coupon.DownVotes, pq.Array(coupon.Categories),
		pq.Array(coupon.Tags), pq.Array(coupon.Regions),
		coupon.StoreType,
	).Scan(&coupon.ID, &coupon.CreatedAt, &coupon.MaterializedScore)
}

func (r *CouponRepository) GetByID(ctx context.Context, id int64) (*models.Coupon, error) {
	const query = `
        SELECT 
            id, created_at, code, title, description,
            discount_value, discount_type, merchant_name, merchant_url,
            start_date, end_date, terms_conditions,
            minimum_purchase_amount, maximum_discount_amount,
            up_votes, down_votes, categories, tags,
            regions, store_type, materialized_score,
            last_score_update
        FROM coupons 
        WHERE id = $1`

	coupon := &models.Coupon{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&coupon.ID, &coupon.CreatedAt, &coupon.Code,
		&coupon.Title, &coupon.Description, &coupon.DiscountValue,
		&coupon.DiscountType, &coupon.MerchantName, &coupon.MerchantURL,
		&coupon.StartDate, &coupon.EndDate, &coupon.TermsConditions,
		&coupon.MinimumPurchaseAmount, &coupon.MaximumDiscountAmount,
		&coupon.UpVotes, &coupon.DownVotes,
		pq.Array(&coupon.Categories), pq.Array(&coupon.Tags),
		pq.Array(&coupon.Regions), &coupon.StoreType,
		&coupon.MaterializedScore, &coupon.LastScoreUpdate,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return coupon, err
}

func (r *CouponRepository) BatchAddVotes(ctx context.Context, votes []models.Vote, voteType string) error {
	const upQuery = `
        UPDATE coupons AS c
        SET up_votes = CASE 
            WHEN v.id = c.id THEN array_append(c.up_votes, v.timestamp)
            ELSE c.up_votes
        END
        FROM (SELECT unnest($1::bigint[]) AS id, unnest($2::timestamp[]) AS timestamp) AS v
        WHERE c.id = v.id`

	const downQuery = `
        UPDATE coupons AS c
        SET down_votes = CASE 
            WHEN v.id = c.id THEN array_append(c.down_votes, v.timestamp)
            ELSE c.down_votes
        END
        FROM (SELECT unnest($1::bigint[]) AS id, unnest($2::timestamp[]) AS timestamp) AS v
        WHERE c.id = v.id`

	ids := make([]int64, len(votes))
	timestamps := make([]time.Time, len(votes))
	for i, v := range votes {
		ids[i] = v.ID
		timestamps[i] = v.Timestamp
	}

	query := upQuery
	if voteType == "down" {
		query = downQuery
	}

	result, err := r.db.ExecContext(ctx, query, pq.Array(ids), pq.Array(timestamps))
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// --- Search and Filtering ---

// SortBy represents the available sorting options for coupons
type SortBy string

const (
	SortByNewest    SortBy = "newest"
	SortByOldest    SortBy = "oldest"
	SortByHighScore SortBy = "high_score"
	SortByLowScore  SortBy = "low_score"
)

// SearchParams contains all parameters for searching and filtering coupons
type SearchParams struct {
	SearchString string
	SortBy       SortBy
	Limit        int
	Offset       int
}

func (r *CouponRepository) Search(ctx context.Context, params SearchParams) ([]*models.Coupon, error) {
	// Base query
	query := `
        SELECT 
            id, created_at, code, title, description,
            discount_value, discount_type, merchant_name, merchant_url,
            start_date, end_date, terms_conditions,
            minimum_purchase_amount, maximum_discount_amount,
            up_votes, down_votes, categories, tags,
            regions, store_type, materialized_score,
            last_score_update
        FROM coupons
        WHERE 1=1
    `

	// Initialize parameters array and counter
	queryParams := make([]interface{}, 0)
	paramCounter := 1

	// Add search condition if search string is provided
	if params.SearchString != "" {
		query += fmt.Sprintf(`
            AND (
                code ILIKE $%d OR
                title ILIKE $%d OR
                description ILIKE $%d OR
                merchant_name ILIKE $%d
            )`, paramCounter, paramCounter, paramCounter, paramCounter)
		searchTerm := "%" + params.SearchString + "%"
		queryParams = append(queryParams, searchTerm)
		paramCounter++
	}

	switch params.SortBy {
	case SortByNewest:
		query += ` ORDER BY created_at DESC`
	case SortByOldest:
		query += ` ORDER BY created_at ASC`
	case SortByHighScore:
		query += ` ORDER BY materialized_score  DESC`
	case SortByLowScore:
		query += ` ORDER BY materialized_score  ASC`
	default:
		query += ` ORDER BY created_at DESC`
	}

	// Add pagination
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, paramCounter, paramCounter+1)
	queryParams = append(queryParams, params.Limit, params.Offset)

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}(rows)

	// Parse results
	var coupons []*models.Coupon
	for rows.Next() {
		coupon := &models.Coupon{}
		err := rows.Scan(
			&coupon.ID, &coupon.CreatedAt, &coupon.Code,
			&coupon.Title, &coupon.Description, &coupon.DiscountValue,
			&coupon.DiscountType, &coupon.MerchantName, &coupon.MerchantURL,
			&coupon.StartDate, &coupon.EndDate, &coupon.TermsConditions,
			&coupon.MinimumPurchaseAmount, &coupon.MaximumDiscountAmount,
			&coupon.UpVotes, &coupon.DownVotes,
			pq.Array(&coupon.Categories), pq.Array(&coupon.Tags),
			pq.Array(&coupon.Regions), &coupon.StoreType,
			&coupon.MaterializedScore, &coupon.LastScoreUpdate,
		)
		if err != nil {
			return nil, err
		}
		coupons = append(coupons, coupon)
	}

	return coupons, nil
}

// GetTotalCount returns the total number of coupons matching the search criteria
// This is useful for pagination
func (r *CouponRepository) GetTotalCount(ctx context.Context, params SearchParams) (int64, error) {
	query := `SELECT COUNT(*) FROM coupons WHERE 1=1`
	queryParams := make([]interface{}, 0)
	paramCounter := 1

	if params.SearchString != "" {
		query += fmt.Sprintf(`
            AND (
                code ILIKE $%d OR
                title ILIKE $%d OR
                description ILIKE $%d OR
                merchant_name ILIKE $%d
            )`, paramCounter, paramCounter, paramCounter, paramCounter)
		searchTerm := "%" + params.SearchString + "%"
		queryParams = append(queryParams, searchTerm)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, queryParams...).Scan(&count)
	return count, err
}

// --- Merchants ---

func (r *CouponRepository) GetMerchants(ctx context.Context) (*models.MerchantResponse, error) {
	query := `SELECT DISTINCT merchant_name, merchant_url FROM coupons ORDER BY merchant_name;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}(rows)

	var merchants []models.Merchant
	for rows.Next() {
		merchant := models.Merchant{}
		err := rows.Scan(&merchant.Name, &merchant.URL)
		if err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}

	merchantResponse := &models.MerchantResponse{
		Total: len(merchants),
		Data:  merchants,
	}

	return merchantResponse, nil
}

// --- Categories ---

func (r *CouponRepository) GetCategories(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT unnest(categories) FROM coupons ORDER BY 1;`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatalf("Failed to close rows: %v", err)
		}
	}(rows)

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
