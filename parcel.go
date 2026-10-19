package main

import (
	"database/sql"
	"errors"
)

var ErrParcelNotFoundOrNotEditable = errors.New("посылка не найдена или её статус не позволяет изменение")

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrParcelNotFoundOrNotEditable
	}
	return nil
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) "+
			"VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel

	err := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
		 FROM parcel
		 WHERE number = ?`,
		number,
	).Scan(
		&p.Number,
		&p.Client,
		&p.Status,
		&p.Address,
		&p.CreatedAt,
	)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
		 FROM parcel
		 WHERE client = ?`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel

	for rows.Next() {
		var p Parcel

		if err := rows.Scan(
			&p.Number,
			&p.Client,
			&p.Status,
			&p.Address,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}

		parcels = append(parcels, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		`UPDATE parcel
		 SET status = ?
		 WHERE number = ?`,
		status, number,
	)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec(
		`UPDATE parcel
		 SET address = ?
		 WHERE number = ? AND status = ?`,
		address, number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec(
		`DELETE FROM parcel
		 WHERE number = ? AND status = ?`,
		number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}
	return checkAffected(res)
}
