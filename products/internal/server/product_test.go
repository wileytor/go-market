package server
/*
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	//"github.com/lahnasti/go-market/internal/models"
	//"github.com/lahnasti/go-market/internal/server/responses"
	//"github.com/lahnasti/go-market/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//mockery --dir=internal/repository --name=Repository --output=mocks/ --outpkg=mocks

func TestGetAllProductsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := new(mocks.Repository)
	srv := &Server{
		Db:    m,
		log:   zerolog.New(os.Stdout),
		Valid: validator.New(),
	}
	r := gin.Default()
	r.GET("/products", srv.GetAllProductsHandler)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	type want struct {
		code     int
		products string
	}
	type test struct {
		name    string
		request string
		method  string
		product []models.Product
		err     error
		want    want
	}
	tests := []test{
		{
			name:    "Test 'GetAllProductsHandler' #1; Default call",
			request: "/products",
			method:  http.MethodGet,
			product: []models.Product{
				{UID: 1, Name: "apple", Description: "fruit", Price: 100, Quantity: 10},
				{UID: 2, Name: "banana", Description: "fruit", Price: 50, Quantity: 5},
			},
			want: want{
				code:     http.StatusOK,
				products: `{"message":"List of products","status":200,"data":[{"uid":1,"name":"apple","description":"fruit","price":100,"quantity":10,"delete":false},{"uid":2,"name":"banana","description":"fruit","price":50,"quantity":5,"delete":false}]}`,
			},
		},
		{
			name:    "Test 'GetAllProductsHandler' #2; Error call",
			request: "/products",
			method:  http.MethodGet,
			err:     errors.New("db error"),
			want: want{
				code:     http.StatusInternalServerError,
				products: `{"status": 500, "message": "Failed to retrieve products", "error": "db error"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				// Устанавливаем поведение мока для ошибки
				m.On("GetAllProducts").Return(nil, tt.err).Once()
			} else {
				// Устанавливаем поведение мока для успешного вызова
				m.On("GetAllProducts").Return(tt.product, nil).Once()
			}

			req, _ := http.NewRequest(tt.method, tt.request, nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			assert.Equal(t, tt.want.code, resp.Code)
			if tt.want.products != "" {
				assert.JSONEq(t, tt.want.products, resp.Body.String())
			}

			m.AssertExpectations(t)
		})
	}
}

func TestGetProductByIDHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := new(mocks.Repository)
	srv := &Server{
		Db:    m,
		log:   zerolog.New(os.Stdout),
		Valid: validator.New(),
	}
	r := gin.Default()
	r.GET("/products/:id", srv.GetProductByIDHandler)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	type want struct {
		code     int
		products string
	}
	type test struct {
		name    string
		request string
		method  string
		product models.Product
		err     error
		want    want
	}
	tests := []test{
		{
			name:    "Test 'GetProductByIDHandler' #1; Default call",
			request: "/products/1",
			method:  http.MethodGet,
			product: models.Product{UID: 1, Name: "apple", Description: "fruit", Price: 100, Quantity: 10, Delete: false},
			want: want{
				code:     http.StatusOK,
				products: `{"message":"Product found","status":200,"data":{"uid":1,"name":"apple","description":"fruit","price":100,"quantity":10,"delete":false}}`,
			},
		},
		{
			name:    "Test 'GetProductByIDHandler' #2; Invalid ID",
			request: "/products/1a",
			method:  http.MethodGet,
			err:     errors.New("invalid id"),
			want: want{
				code:     http.StatusBadRequest,
				products: `{"status": 400, "message": "Invalid id", "error": "invalid id"}`,
			},
		},
		{
			name:    "Test 'GetProductByIDHandler' #3; Product not found",
			request: "/products/100",
			method:  http.MethodGet,
			err:     responses.ErrNotFound,
			want: want{
				code:     http.StatusInternalServerError,
				products: `{"status": 500, "message": "Product not found", "error": "product not found"}`,
			},
		},
		{
			name:    "Test 'GetProductByIDHandler' #4; Other error",
			request: "/product/100",
			method:  http.MethodGet,
			err:     errors.New("some other error"),
			want: want{
				code:     http.StatusNotFound,
				products: `{"status": 404, "message": "error", "error": "some other error"}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				m.On("GetProductByID", mock.Anything).Return(models.Product{}, tt.err)
			} else {
				m.On("GetProductByID", mock.Anything).Return(tt.product, nil).Once()
			}

			req, _ := http.NewRequest(tt.method, tt.request, nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			assert.Equal(t, tt.want.code, resp.Code)
			if tt.want.products != "" {
				assert.JSONEq(t, tt.want.products, tt.want.products, resp.Body.String())
			}
			m.AssertExpectations(t)
		})
	}
}

func TestAddProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := new(mocks.Repository)
	srv := &Server{
		Db:    m,
		log:   zerolog.New(os.Stdout),
		Valid: validator.New(),
	}
	r := gin.Default()
	r.POST("/products/add", srv.AddProductHandler)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	type want struct {
		code     int
		products string
	}
	type test struct {
		name    string
		request string
		method  string
		product string
		err     error
		want    want
	}
	tests := []test{
		{
			name:    "Test 'AddProductHandler' #1; Default call",
			request: "/products/add",
			method:  http.MethodPost,
			product: `{"name":"banana","description":"fruit","price":50,"quantity":1}`,
			want: want{
				code:     http.StatusCreated,
				products: `{"data":1, "message":"Product added", "status":201}`,
			},
		},
		{
			name:    "Test 'AddProductHandler' #2; Invalid request data",
			request: "/products/add",
			method:  http.MethodPost,
			product: `{"name":"banana","description":"fruit","price":1,"quantity":1}`,
			err:     errors.New("invalid request data"),
			want: want{
				code:     http.StatusInternalServerError,
				products: `{"error":"invalid request data", "message":"error", "status":500}`,
			},
		},
		{
			name:    "Test 'AddProductHandler' #3; Not a valid product",
			request: "/products/add",
			method:  http.MethodPost,
			product: `{"name":"banana","description":"banana","price":1}`,
			err:     errors.New("not a valid product"),
			want: want{
				code:     http.StatusBadRequest,
				products: `{"error":"Key: 'Product.Quantity' Error:Field validation for 'Quantity' failed on the 'required' tag", "message":"Not a valid product", "status":400}`,
			},
		},
		{
			name:    "Test 'AddProductHandler' #4; Quantity cannot be negative",
			request: "/products/add",
			method:  http.MethodPost,
			product: `{"name":"banana","description":"fruit","price":1,"quantity:0}`,
			err:     errors.New("quantity cannot be negative"),
			want: want{
				code:     http.StatusBadRequest,
				products: `{"error":"unexpected EOF", "message":"Invalid request data", "status":400}`,
			},
		},
		{
			name:    "Test 'AddProductHandler' #5; Product name already exists",
			request: "/products/add",
			method:  http.MethodPost,
			product: `{"name":"banana","description":"fruit","price":1,"quantity:1}`,
			err:     errors.New("product name already exists"),
			want: want{
				code:     http.StatusBadRequest,
				products: `{"error":"unexpected EOF", "message":"Invalid request data", "status":400}`,
			},
		},
		{
			name:    "Test 'AddProductHandler' #6; Invalid request data",
			request: "/products/add",
			method:  http.MethodPost,
			product: `{"name":"banana","description":"fruit","price":1,"quantity:1}`,
			err:     errors.New("error"),
			want: want{
				code:     http.StatusBadRequest,
				products: `{"error":"unexpected EOF", "message":"Invalid request data", "status":400}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var product models.Product
			json.Unmarshal([]byte(tt.product), &product)                     // Десериализация JSON в структуру
			m.On("IsProductUnique", product.Name).Return(tt.err == nil, nil) // true, если ошибки нет

			if tt.err != nil {
				m.On("AddProduct", mock.Anything).Return(0, tt.err) // Возвращаем 0 и ошибку
			} else {
				m.On("AddProduct", product).Return(1, nil)
			}
			req, _ := http.NewRequest(tt.method, tt.request, bytes.NewBufferString(tt.product))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			assert.Equal(t, tt.want.code, resp.Code)
			if tt.want.products != "" {
				assert.JSONEq(t, tt.want.products, resp.Body.String())
			}
		})
	}
}

func TestUpdateProductHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := new(mocks.Repository)
	srv := &Server{
		Db:    m,
		log:   zerolog.New(os.Stdout),
		Valid: validator.New(),
	}
	r := gin.Default()
	r.PUT("/products/:id", srv.UpdateProductHandler)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	type want struct {
		code   int
		answer string
	}
	type test struct {
		name    string
		request string
		method  string
		product string
		err     error
		want    want
	}
	tests := []test{
		{
			name:    "Test 'UpdateProductHandler' #1; Default call",
			request: "/products/1",
			method:  http.MethodPut,
			product: `{"name":"apple","description":"fruit","price":50,"quantity":1}`,
			want: want{
				code:   http.StatusOK,
				answer: `{"data":1, "message":"Product updated", "status":200}`,
			},
		},
		{
			name:    "Test 'UpdateProductHandler' #2; Invalid request data",
			request: "/products/1",
			method:  http.MethodPut,
			product: `{"name":"apple","description":"fruit","price":1,"quantity":1}`,
			err:     errors.New("invalid request data"),
			want: want{
				code:   http.StatusInternalServerError,
				answer: `{"error":"invalid request data", "message":"error", "status":500}`,
			},
		},
		{
			name:    "Test 'UpdateProductHandler' #3; Not a valid product",
			request: "/products/1",
			method:  http.MethodPut,
			product: `{"name":"apple","description":"banana","price":1}`,
			err:     errors.New("not a valid product"),
			want: want{
				code:   http.StatusBadRequest,
				answer: `{"error":"Key: 'Product.Quantity' Error:Field validation for 'Quantity' failed on the 'required' tag", "message":"Not a valid product", "status":400}`,
			},
		},
		{
			name:    "Test 'UpdateProductHandler' #4; Error other",
			request: "/products/1",
			method:  http.MethodPut,
			product: `{"name":"apple","description":"fruit","price":1,"quantity":1}`,
			err:     errors.New("error"),
			want: want{
				code:   http.StatusInternalServerError,
				answer: `{"error":"invalid request data", "message":"error", "status":500}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var product models.Product
			product.UID = 1
			err := json.Unmarshal([]byte(tt.product), &product) // Десериализация JSON в структуру
			if err != nil {
				t.Fatalf("failed to unmarshal product: %v", err)
			}

			if tt.err != nil {
				m.On("UpdateProduct", product.UID, product).Return(0, tt.err) // Возвращаем ошибку
			} else {
				m.On("UpdateProduct", product.UID, product).Return(product.UID, nil).Once()
			}
			req, _ := http.NewRequest(tt.method, tt.request, bytes.NewBufferString(tt.product))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)

			assert.Equal(t, tt.want.code, resp.Code)
			if tt.want.answer != "" {
				assert.JSONEq(t, tt.want.answer, resp.Body.String())
			}
		})
	}
}

func TestDeleteProductHandler(t *testing.T) {
		gin.SetMode(gin.TestMode)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := new(mocks.Repository)
		deleteChan := make(chan int, 1)
		errorChan := make(chan error, 1)


		srv := &Server{
			Db: m,
			deleteChan: deleteChan,
			log:   zerolog.New(os.Stdout),
			ErrorChan: errorChan,
		}
		r := gin.Default()
		r.DELETE("/products/:id", srv.DeleteProductHandler)
		httpSrv := httptest.NewServer(r)
		defer httpSrv.Close()

		type want struct {
			code int
			answer string
		}

		type test struct {
			name string
			request string
			method string
			err error
			want want
		}
		tests := []test{
			{
				name:    "Test 'DeleteProductHandler' #1; Default call",
				request: "/products/1",
				method:  http.MethodDelete,
				want: want{
					code:   http.StatusOK,
					answer: `{"data":1, "message":"Product deleted", "status":200}`,
				},
			},
			{
				name:    "Test 'DeleteProductHandler' #2; Invalid ID format",
				request: "/products/abc",
				method:  http.MethodDelete,
				want: want{
					code:   http.StatusBadRequest,
					answer: `{"error":"strconv.Atoi: parsing \"abc\": invalid syntax", "message":"Invalid id", "status":400}`,
				},
			},
			{
				name:    "Test 'DeleteProductHandler' #3; Database error",
				request: "/products/1",
				method:  http.MethodDelete,
				err:     errors.New("database error"),
				want: want{
					code:   http.StatusInternalServerError,
					answer: `{"error":"database error", "message":"error", "status":500}`,
				},
			},
		}
	
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Мокаем поведение метода SetDeleteStatus
				if tt.err != nil {
					m.On("SetDeleteStatus", 1).Return(tt.err)
				} else {
					m.On("SetDeleteStatus", 1).Return(nil)
				}
	
				req, _ := http.NewRequest(tt.method, tt.request, nil)
				req.Header.Set("Content-Type", "application/json")
				resp := httptest.NewRecorder()
	
				r.ServeHTTP(resp, req)
	
				assert.Equal(t, tt.want.code, resp.Code)
				if tt.want.answer != "" {
					assert.JSONEq(t, tt.want.answer, resp.Body.String())
				}
	
				// Проверяем выполнение моков
				m.AssertExpectations(t)
			})
		}
	}

func TestDeleter(t *testing.T) {
	// Мокируем базу данных
	m := new(mocks.Repository)
	deleteChan := make(chan int, 10)
	errorChan := make(chan error, 1)

	// Создаем сервер с моком, каналами и логером
	srv := &Server{
		Db:         m,
		deleteChan: deleteChan,
		ErrorChan:  errorChan,
		log:        zerolog.New(os.Stdout),
	}

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Добавляем тестовые данные в deleteChan
	for i := 1; i <= 5; i++ {
		deleteChan <- i
	}

	// Мокаем вызов DeleteProducts для успешного удаления
	m.On("DeleteProducts").Return(nil)

	// Запускаем deleter в горутине
	go srv.deleter(ctx)

	// Даем время горутине для обработки
	time.Sleep(2 * time.Second)

	// Проверяем, что метод DeleteProducts был вызван
	m.AssertExpectations(t)

	// Проверяем, что канал очистился
	assert.Equal(t, 0, len(srv.deleteChan))

	// Проверяем, что ошибок не было
	select {
	case err := <-errorChan:
		t.Fatalf("Expected no errors, but got: %v", err)
	default:
		// Ожидаем, что канал ошибок будет пуст
	}
}
*/
