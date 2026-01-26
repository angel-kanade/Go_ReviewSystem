package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v1 "review-service/api/review/v1"
	"review-service/internal/data/model"
	"review-service/pkg/snowflake"
)

type ReviewRepo interface {
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	GetReviewByOrderID(context.Context, int64) ([]*model.ReviewInfo, error)
}

type ReviewUsecase struct {
	repo ReviewRepo
}

// NewReviewUsecase new a Review usecase.
func NewReviewUsecase(repo ReviewRepo) *ReviewUsecase {
	return &ReviewUsecase{
		repo: repo,
	}
}

func (uc *ReviewUsecase) CreateReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	log.Infof("[biz] CreateReview, req: %v", review)
	// 数据校验
	reviews, err := uc.repo.GetReviewByOrderID(ctx, review.OrderID)
	if err != nil {
		log.Debugf("[biz] CreateReview Error, 该订单评价已存在")
		return nil, v1.ErrorDbFailed("查询数据库失败")
	}
	if len(reviews) > 0 {
		log.Debugf("[biz] CreateReview Error, 该订单评价已存在")
		return nil, v1.ErrorOrderReviewed("订单：%d 已评价", review.OrderID)
	}
	// 生成reviewID (Snowflake)
	review.ReviewID = snowflake.GetID()
	// 查询订单和商品快照信息
	// 实际业务场景下就需要查询订单服务和商家服务（使用RPC）
	// 拼装数据入库
	return uc.repo.SaveReview(ctx, review)
}
