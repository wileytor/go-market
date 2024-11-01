package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wileytor/go-market/common/models"
	mocks "github.com/wileytor/go-market/mocks_prod"
	"github.com/wileytor/go-market/products/internal/logger"
	"github.com/wileytor/go-market/products/internal/server/responses"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllProductsHandler(t *testing.T) {
	var srv Server
	r := gin.Default()
	r.GET("/products", srv.GetAllProductsHandler)
	httpSrv := httptest.NewServer(r)

	type want struct {
		errFlag    bool
		statusCode int
		products   string
	}
	type test struct {
		name     string
		method   string
		request  string
		products []models.Product
		err      error
		want     want
	}

	tests := []test{
		{
			name:    "Test GetAllProductsHandler; Case 1:",
			method:  http.MethodGet,
			request: "/products",
			err:     nil,
			products: []models.Product{
				{
					UID:         1,
					Name:        "apple",
					Description: "red",
					Price:       10,
					Delete:      false,
					Quantity:    12,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				products:   `{"status":200,"message":"List of products","data":[{"uid":1,"name":"apple","description":"red","price":10,"delete":false,"quantity":12}]}`,
				errFlag:    false,
			},
		},
		{
			name:     "Test GetAllProductsHandler; Case 2;",
			method:   http.MethodGet,
			request:  "/products",
			err:      fmt.Errorf("test error"),
			products: nil,
			want: want{
				statusCode: http.StatusInternalServerError,
				products:   `{"status":500,"message":"Failed to retrieve products","error":"test error"}`,
				errFlag:    true,
			},
		},
	}

	log := logger.SetupLogger(true)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockRepository(ctrl)
			defer ctrl.Finish()
			m.EXPECT().GetAllProducts().Return(tc.products, tc.err)
			srv.Db = m
			req := resty.New().R()
			req.Method = tc.method
			req.URL = httpSrv.URL + tc.request
			resp, err := req.Send()
			log.Debug().Err(err).Str("body", string(resp.Body())).Any("str", resp.String()).Send()
			if !tc.want.errFlag {
				assert.NoError(t, err)
			}
			assert.Equal(t, resp.StatusCode(), tc.want.statusCode)
			assert.Equal(t, tc.want.products, string(resp.Body()))
		})
	}
}

func TestGetProductByIDHandler(t *testing.T) {
	var srv Server
	r := gin.Default()
	r.GET("/products/:id", srv.GetProductByIDHandler)
	httpSrv := httptest.NewServer(r)

	type want struct {
		errFlag    bool
		statusCode int
		product    string
	}
	type test struct {
		name    string
		method  string
		request string
		product models.Product
		err     error
		want    want
	}

	tests := []test{
		{
			name:    "Test GetProductByIDHandler; Case 1;",
			method:  http.MethodGet,
			request: "/products/1",
			err:     nil,
			product: models.Product{
				UID:         1,
				Name:        "apple",
				Description: "red",
				Price:       10,
				Delete:      false,
				Quantity:    12,
			},
			want: want{
				statusCode: http.StatusOK,
				product:    `{"status":200,"message":"Product found","data":{"uid":1,"name":"apple","description":"red","price":10,"delete":false,"quantity":12}}`,
				errFlag:    false,
			},
		},
		{
			name:    "Test GetProductByIDHandler; Case 2;",
			method:  http.MethodGet,
			request: "/products/a",
			err:     fmt.Errorf("strconv.Atoi: parsing \"a\": invalid syntax"),
			product: models.Product{},
			want: want{
				statusCode: http.StatusBadRequest,
				product:    `{"status":400,"message":"Invalid id","error":"strconv.Atoi: parsing \"a\": invalid syntax"}`,
				errFlag:    true,
			},
		},
		{
			name:    "Test GetProductByIDHandler; Case 3;",
			method:  http.MethodGet,
			request: "/products/1",
			err:     responses.ErrNotFound,
			product: models.Product{},
			want: want{
				statusCode: http.StatusNotFound,
				product:    `{"status":404,"message":"Product not found","error":"product not found"}`,
				errFlag:    true,
			},
		},
		{
			name:    "Test GetProductByIDHandler; Case 4;",
			method:  http.MethodGet,
			request: "/products/3",
			err:     fmt.Errorf("some internal error"),
			product: models.Product{},
			want: want{
				statusCode: http.StatusInternalServerError,
				product:    `{"status":500,"message":"Failed to retrieve product","error":"some internal error"}`,
				errFlag:    true,
			},
		},
	}
	log := logger.SetupLogger(true)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockRepository(ctrl)
			defer ctrl.Finish()
			if !tc.want.errFlag {
				m.EXPECT().GetProductByID(1).Return(tc.product, tc.err)
			} else {
				m.EXPECT().GetProductByID(mock.Anything).Return(models.Product{}, tc.err)
			}
			srv.Db = m
			req := resty.New().R()
			req.Method = tc.method
			req.URL = httpSrv.URL + tc.request
			resp, err := req.Send()
			log.Debug().Err(err).Str("body", string(resp.Body())).Any("str", resp.String()).Send()
			if !tc.want.errFlag {
				assert.NoError(t, err)
			}
			assert.Equal(t, resp.StatusCode(), tc.want.statusCode)
			assert.Equal(t, tc.want.product, string(resp.Body()))
		})
	}
}

func TestAddProductHandler(t *testing.T) {

}
