package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	//"github.com/lahnasti/go-market/mocks"
)

func TestMakePurchaseHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//m := new(mocks.Repository)
	//srv := &Server{
	//        Db:    m,

	//  }

}
