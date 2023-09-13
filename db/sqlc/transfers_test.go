package db

import (
	"context"
	"testing"

	"github.com/aulas/demo-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T) Transfer {
	fromAcc := createRandomAccount(t)
	toAcc := createRandomAccount(t)
	amount := util.RandomMoney()
	arg := CreateTransferParams{
		FromAccountID: fromAcc.ID,
		ToAccountID:   toAcc.ID,
		Amount:        amount,
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, transfer.FromAccountID, fromAcc.ID)
	require.Equal(t, transfer.ToAccountID, toAcc.ID)
	require.Equal(t, transfer.Amount, amount)
	require.NotZero(t, transfer.ID)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	createRandomTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	trans1 := createRandomTransfer(t)
	trans2, err := testQueries.GetTransfer(context.Background(), trans1.ID)
	require.NoError(t, err)

	require.Equal(t, trans2.FromAccountID, trans1.FromAccountID)
	require.Equal(t, trans2.ToAccountID, trans1.ToAccountID)
	require.Equal(t, trans2.Amount, trans1.Amount)
	require.Equal(t, trans2.Amount, trans1.Amount)
}

func TestListTransfer(t *testing.T) {
	for i := 0; i < 5; i++ {
		createRandomTransfer(t)
	}

	arg := ListTransferParams{
		Limit:  5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}

}
