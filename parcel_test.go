package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) ParcelStore {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	db.SetMaxOpenConns(1)

	_, err = db.Exec(`
		CREATE TABLE parcel (
			number INTEGER PRIMARY KEY AUTOINCREMENT,
			client INTEGER NOT NULL,
			status TEXT NOT NULL,
			address TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return NewParcelStore(db)
}

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	store := newTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, storedParcel)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestSetAddress(t *testing.T) {
	store := newTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newAddress := "new test address"

	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, storedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	store := newTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	newStatus := ParcelStatusSent

	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, storedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	store := newTestStore(t)

	client := 12345
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	expected := make(map[int]Parcel)

	for i := range parcels {
		parcels[i].Client = client

		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		parcels[i].Number = id
		expected[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for _, parcel := range storedParcels {
		expectedParcel, exists := expected[parcel.Number]
		require.True(t, exists)
		require.Equal(t, expectedParcel, parcel)
	}
}

func TestDeleteSentParcel(t *testing.T) {
	store := newTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	err = store.Delete(id)
	require.ErrorIs(t, err, ErrParcelNotFoundOrNotEditable)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, storedParcel.Status)
}

func TestSetAddressSentParcel(t *testing.T) {
	store := newTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	err = store.SetAddress(id, "another address")
	require.ErrorIs(t, err, ErrParcelNotFoundOrNotEditable)

	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Address, storedParcel.Address)
}
