package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/firzatullahd/cats-social-api/internal/entity"
	"github.com/firzatullahd/cats-social-api/internal/model"
	error_envelope "github.com/firzatullahd/cats-social-api/internal/model/error"
	"github.com/firzatullahd/cats-social-api/internal/utils/constant"
	"github.com/firzatullahd/cats-social-api/internal/utils/logger"
)

func (u *Usecase) CreateCat(ctx context.Context, in *model.CreateCatRequest, userId uint64) (*model.CreateCatResponse, error) {
	logCtx := fmt.Sprintf("%T.CreateCat", u)
	var err error

	inputRegister, err := validateRegisterCat(in)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return nil, err
	}

	tx, err := u.repo.WithTransaction()
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	catId, err := u.repo.CreateCat(ctx, tx, inputRegister)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return nil, err
	}
	tx.Commit()

	cats, err := u.repo.FindCat(ctx, &model.FilterFindCat{ID: []uint64{catId}})
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return nil, err
	}

	return &model.CreateCatResponse{
		CreatedAt: cats[0].CreatedAt.Format(constant.DefaultDateFormat),
		ID:        fmt.Sprintf("%v", cats[0].ID),
	}, nil
}

func validateRegisterCat(in *model.CreateCatRequest) (*entity.Cat, error) {
	var err error

	if len(in.Name) < 1 || len(in.Name) > 30 {
		return nil, error_envelope.ErrValidation
	}

	if len(in.Sex) == 0 {
		return nil, error_envelope.ErrValidation
	}

	tSex, err := entity.StringToSex(in.Sex)
	if err != nil {
		return nil, error_envelope.ErrValidation
	}

	if len(in.Race) == 0 {
		return nil, error_envelope.ErrValidation
	}

	tRace, err := entity.StringToRace(in.Race)
	if err != nil {
		return nil, error_envelope.ErrValidation
	}

	if in.AgeInMonth <= 0 {
		return nil, error_envelope.ErrValidation
	}

	if len(in.Description) <= 1 || len(in.Description) > 200 {
		return nil, error_envelope.ErrValidation
	}

	if len(in.ImageUrls) <= 0 {
		return nil, error_envelope.ErrValidation
	}

	return &entity.Cat{
		UserID:      in.UserID,
		Name:        in.Name,
		Sex:         tSex,
		Race:        tRace,
		ImageUrls:   in.ImageUrls,
		Age:         in.AgeInMonth,
		Description: in.Description,
	}, nil
}

func (u *Usecase) DeleteCat(ctx context.Context, catId, userId uint64) error {
	logCtx := fmt.Sprintf("%T.DeleteCat", u)
	var err error

	tx, err := u.repo.WithTransaction()
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = u.repo.DeleteCat(ctx, tx, catId, userId)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return err
	}

	tx.Commit()

	return nil
}

func (u *Usecase) UpdateCat(ctx context.Context, in *model.UpdateCatRequest) error {
	logCtx := fmt.Sprintf("%T.UpdateCat", u)
	var err error

	updateInput, err := validateUpdateCat(in)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return err
	}

	tx, err := u.repo.WithTransaction()
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if updateInput.Sex != nil {
		matches, err := u.repo.FindMatch(ctx, &model.FilterFindMatch{
			CatId: []uint64{in.ID},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, logCtx, err)
			return err
		}

		if len(matches) > 0 {
			return error_envelope.ErrEmailExists
		}
	}

	err = u.repo.UpdateCat(ctx, tx, updateInput)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return err
	}

	tx.Commit()

	return nil
}

func validateUpdateCat(in *model.UpdateCatRequest) (*model.InputUpdateCat, error) {
	var updateCat model.InputUpdateCat

	updateCat.ID = in.ID
	updateCat.UserID = in.UserID

	if in.Name != nil {
		if len(*in.Name) <= 1 || len(*in.Name) > 30 {
			return nil, error_envelope.ErrValidation
		}
		updateCat.Name = in.Name
	}

	if in.Sex != nil {
		t, err := entity.StringToSex(*in.Sex)
		if err != nil {
			return &updateCat, error_envelope.ErrValidation
		}

		*updateCat.Sex = t.String()
	}

	if in.Race != nil {
		t, err := entity.StringToRace(*in.Race)
		if err != nil {
			return &updateCat, error_envelope.ErrValidation
		}

		*updateCat.Race = t.String()
	}

	if in.ImageUrls != nil || len(in.ImageUrls) > 0 {
		updateCat.ImageUrls = in.ImageUrls
	}

	if in.AgeInMonth != nil {
		updateCat.Age = in.AgeInMonth
	}

	if in.Description != nil {
		if len(*in.Description) <= 1 || len(*in.Description) > 200 {
			return nil, error_envelope.ErrValidation
		}
		updateCat.Description = in.Description
	}

	return nil, nil
}

func (u *Usecase) FindCat(ctx context.Context, in *model.FindCatRequest) ([]model.FindCatResponse, error) {
	logCtx := fmt.Sprintf("%T.FindCat", u)
	var err error

	filter, err := parseFilterFindCat(in)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return nil, error_envelope.ErrValidation
	}

	cats, err := u.repo.FindCat(ctx, filter)
	if err != nil {
		logger.Error(ctx, logCtx, err)
		return nil, err
	}

	var resp []model.FindCatResponse
	for _, cat := range cats {
		resp = append(resp, model.FindCatResponse{
			ID:          fmt.Sprintf("%v", cat.ID),
			Name:        cat.Name,
			Sex:         cat.Sex.String(),
			Race:        cat.Race.String(),
			ImageUrls:   cat.ImageUrls,
			AgeInMonth:  cat.Age,
			Description: cat.Description,
			HasMatched:  cat.HasMatched,
			CreatedAt:   cat.CreatedAt.Format(constant.DefaultDateFormat),
		})
	}

	return resp, nil
}

func parseFilterFindCat(in *model.FindCatRequest) (*model.FilterFindCat, error) {
	out := new(model.FilterFindCat)

	limit, err := strconv.Atoi(in.Limit)
	if err != nil {
		return nil, err
	}
	out.Limit = limit

	offset, err := strconv.Atoi(in.Offset)
	if err != nil {
		return nil, err
	}
	out.Offset = offset

	if in.HasMatched != "" {
		hasMatched, err := strconv.ParseBool(in.HasMatched)
		if err != nil {
			return nil, err
		}

		out.HasMatched = &hasMatched
	}

	if in.Owned != "" {
		owned, err := strconv.ParseBool(in.Owned)
		if err != nil {
			return nil, err
		}

		if owned {
			out.UserID = &in.UserId
		}
	}

	if in.Sex != "" {
		_, err := entity.StringToSex(in.Sex)
		if err != nil {
			return nil, err
		}

		out.Sex = &in.Sex
	}

	if in.Race != "" {
		_, err := entity.StringToRace(in.Race)
		if err != nil {
			return nil, err
		}

		out.Race = &in.Race
	}

	if in.ID != "" {
		id, err := strconv.ParseUint(in.ID, 10, 64)
		if err != nil {
			return nil, err
		}
		out.ID = []uint64{id}
	}

	if in.SearchName != "" {
		out.SearchName = &in.SearchName
	}

	if in.Age != "" {
		ok := false
		allowed := []string{">4", "<4", "4"}
		for _, v := range allowed {
			if in.Age == v {
				ok = true
				break
			}
		}

		if !ok {
			return nil, error_envelope.ErrValidation
		}

		out.Age = &in.Age
	}

	return out, nil
}
