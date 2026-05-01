package repository

import (
	"context"
	"database/sql"
	"go-banking-service/internal/database"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/model"
)

type PaymentScheduleRepository interface {
	CreateBatch(ctx context.Context, schedules []*model.PaymentSchedule) error
	GetUnpaidDue(ctx context.Context, dueDate string) ([]*model.PaymentSchedule, error)
	UpdatePaidStatus(ctx context.Context, id string, isPaid bool, lateFeeApplied bool) error
	GetByCreditID(ctx context.Context, creditID string) ([]*model.PaymentSchedule, error)
	UpdatePaidStatusTx(ctx context.Context, tx *sql.Tx, id string, isPaid bool, lateFeeApplied bool) error
}

type paymentScheduleRepository struct {
	storage *database.Storage
}

func NewPaymentScheduleRepository(storage *database.Storage) PaymentScheduleRepository {
	return &paymentScheduleRepository{storage: storage}
}

func (r *paymentScheduleRepository) CreateBatch(ctx context.Context, schedules []*model.PaymentSchedule) error {
	tx, err := r.storage.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction for batch insert", "error", err)
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO payment_schedules (id, credit_id, payment_date, amount, is_paid, late_fee_applied) VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		logger.Error("Failed to prepare statement for batch insert", "error", err)
		return err
	}
	defer stmt.Close()

	for _, schedule := range schedules {
		_, err := stmt.ExecContext(ctx, schedule.ID, schedule.CreditID, schedule.PaymentDate, schedule.Amount, schedule.IsPaid, schedule.LateFeeApplied)
		if err != nil {
			logger.Error("Failed to execute batch insert for schedule", "error", err, "schedule_id", schedule.ID)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit batch insert transaction", "error", err)
		return err
	}
	return nil
}

func (r *paymentScheduleRepository) GetUnpaidDue(ctx context.Context, dueDate string) ([]*model.PaymentSchedule, error) {
	rows, err := r.storage.DB.QueryContext(ctx, `SELECT id, credit_id, payment_date, amount, is_paid, late_fee_applied FROM payment_schedules WHERE payment_date <= $1 AND is_paid = FALSE`, dueDate)
	if err != nil {
		logger.Error("Failed to get unpaid schedules from DB", "error", err, "due_date", dueDate)
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.PaymentSchedule
	for rows.Next() {
		var schedule model.PaymentSchedule
		if err := rows.Scan(&schedule.ID, &schedule.CreditID, &schedule.PaymentDate, &schedule.Amount, &schedule.IsPaid, &schedule.LateFeeApplied); err != nil {
			logger.Error("Failed to scan payment schedule row", "error", err)
			return nil, err
		}
		schedules = append(schedules, &schedule)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over payment schedule rows", "error", err)
		return nil, err
	}
	return schedules, nil
}

func (r *paymentScheduleRepository) UpdatePaidStatus(ctx context.Context, id string, isPaid bool, lateFeeApplied bool) error {
	query := `UPDATE payment_schedules SET is_paid = $1, late_fee_applied = $2 WHERE id = $3`
	_, err := r.storage.DB.ExecContext(ctx, query, isPaid, lateFeeApplied, id)
	if err != nil {
		logger.Error("Failed to update paid status in DB", "error", err, "schedule_id", id, "is_paid", isPaid, "late_fee_applied", lateFeeApplied)
		return err
	}
	return nil
}

func (r *paymentScheduleRepository) GetByCreditID(ctx context.Context, creditID string) ([]*model.PaymentSchedule, error) {
	rows, err := r.storage.DB.QueryContext(ctx, `SELECT id, credit_id, payment_date, amount, is_paid, late_fee_applied FROM payment_schedules WHERE credit_id = $1 ORDER BY payment_date ASC`, creditID)
	if err != nil {
		logger.Error("Failed to get payment schedules by credit ID from DB", "error", err, "credit_id", creditID)
		return nil, err
	}
	defer rows.Close()

	var schedules []*model.PaymentSchedule
	for rows.Next() {
		var schedule model.PaymentSchedule
		if err := rows.Scan(&schedule.ID, &schedule.CreditID, &schedule.PaymentDate, &schedule.Amount, &schedule.IsPaid, &schedule.LateFeeApplied); err != nil {
			logger.Error("Failed to scan payment schedule row", "error", err)
			return nil, err
		}
		schedules = append(schedules, &schedule)
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating over payment schedule rows", "error", err)
		return nil, err
	}
	return schedules, nil
}

func (r *paymentScheduleRepository) UpdatePaidStatusTx(ctx context.Context, tx *sql.Tx, id string, isPaid bool, lateFeeApplied bool) error {
	query := `UPDATE payment_schedules SET is_paid = $1, late_fee_applied = $2 WHERE id = $3`
	_, err := tx.ExecContext(ctx, query, isPaid, lateFeeApplied, id)
	if err != nil {
		logger.Error("Failed to update paid status in DB (tx)", "error", err, "schedule_id", id, "is_paid", isPaid, "late_fee_applied", lateFeeApplied)
		return err
	}
	return nil
}
