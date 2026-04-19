package user

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestService_Upsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("invalid user nil", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		got, err := svc.Upsert(ctx, nil)
		require.Error(t, err)
		require.EqualError(t, err, "invalid user")
		require.Nil(t, got)
	})

	t.Run("invalid user vk id is zero", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		got, err := svc.Upsert(ctx, &User{})
		require.Error(t, err)
		require.EqualError(t, err, "invalid user")
		require.Nil(t, got)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		input := &User{
			UserVKID: 123,
			PeerID:   456,
		}
		expected := &User{
			ID:       1,
			UserVKID: 123,
			PeerID:   456,
		}

		repo.EXPECT().
			Upsert(ctx, input).
			Return(expected, nil).
			Once()

		got, err := svc.Upsert(ctx, input)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		input := &User{
			UserVKID: 123,
		}
		repoErr := errors.New("repo error")

		repo.EXPECT().
			Upsert(ctx, input).
			Return(nil, repoErr).
			Once()

		got, err := svc.Upsert(ctx, input)
		require.Error(t, err)
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, got)
	})
}

func TestService_ReduceCredits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("invalid user nil", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		err := svc.ReduceCredits(ctx, nil)
		require.EqualError(t, err, "invalid user")
	})

	t.Run("invalid user id is zero", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		err := svc.ReduceCredits(ctx, &User{})
		require.EqualError(t, err, "invalid user")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		usr := &User{ID: 42}

		repo.EXPECT().
			ReduceCredits(ctx, int64(42)).
			Return(nil).
			Once()

		err := svc.ReduceCredits(ctx, usr)
		require.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		usr := &User{ID: 42}
		repoErr := errors.New("repo error")

		repo.EXPECT().
			ReduceCredits(ctx, int64(42)).
			Return(repoErr).
			Once()

		err := svc.ReduceCredits(ctx, usr)
		require.Error(t, err)
		require.ErrorIs(t, err, repoErr)
	})
}

func TestService_IncreaseCreditsTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		var tx pgx.Tx

		userID := int64(10)
		credits := int64(5)

		repo.EXPECT().
			IncreaseCreditsTx(ctx, tx, userID, credits).
			Return(nil).
			Once()

		err := svc.IncreaseCreditsTx(ctx, tx, userID, credits)
		require.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		var tx pgx.Tx

		userID := int64(10)
		credits := int64(5)
		repoErr := errors.New("repo error")

		repo.EXPECT().
			IncreaseCreditsTx(ctx, tx, userID, credits).
			Return(repoErr).
			Once()

		err := svc.IncreaseCreditsTx(ctx, tx, userID, credits)
		require.Error(t, err)
		require.ErrorIs(t, err, repoErr)
	})
}

func TestService_FindByVKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		expected := &User{
			ID:       1,
			UserVKID: 123,
		}

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(expected, nil).
			Once()

		got, err := svc.FindByVKID(ctx, 123)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(nil, nil).
			Once()

		got, err := svc.FindByVKID(ctx, 123)
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		repoErr := errors.New("repo error")

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(nil, repoErr).
			Once()

		got, err := svc.FindByVKID(ctx, 123)
		require.Error(t, err)
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, got)
	})
}

func TestService_FindByVkIDOrUpsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("find existing user", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		expected := &User{
			ID:       1,
			UserVKID: 123,
			PeerID:   999,
		}

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(expected, nil).
			Once()

		got, err := svc.FindByVkIDOrUpsert(ctx, 123)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("create user when not found", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		created := &User{
			ID:       1,
			UserVKID: 123,
			PeerID:   0,
		}

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(nil, nil).
			Once()

		repo.EXPECT().
			Upsert(ctx, &User{
				UserVKID: 123,
				PeerID:   0,
			}).
			Return(created, nil).
			Once()

		got, err := svc.FindByVkIDOrUpsert(ctx, 123)
		require.NoError(t, err)
		require.Equal(t, created, got)
	})

	t.Run("find error", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		repoErr := errors.New("repo error")

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(nil, repoErr).
			Once()

		got, err := svc.FindByVkIDOrUpsert(ctx, 123)
		require.Error(t, err)
		require.ErrorContains(t, err, "cannot get user")
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, got)
	})

	t.Run("upsert error", func(t *testing.T) {
		t.Parallel()

		repo := NewMockRepository(t)
		svc := NewService(repo)

		repoErr := errors.New("repo error")

		repo.EXPECT().
			FindByVKID(ctx, int64(123)).
			Return(nil, nil).
			Once()

		repo.EXPECT().
			Upsert(ctx, &User{
				UserVKID: 123,
				PeerID:   0,
			}).
			Return(nil, repoErr).
			Once()

		got, err := svc.FindByVkIDOrUpsert(ctx, 123)
		require.Error(t, err)
		require.ErrorContains(t, err, "cannot create user")
		require.ErrorIs(t, err, repoErr)
		require.Nil(t, got)
	})
}
