package handler

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/OshakbayAigerim/read_space/book_service/internal/domain"
	"github.com/OshakbayAigerim/read_space/book_service/internal/usecase"
	pb "github.com/OshakbayAigerim/read_space/book_service/proto"
)

type BookHandler struct {
	pb.UnimplementedBookServiceServer
	usecase usecase.BookUseCase
	nc      *nats.Conn
}

func NewBookHandler(u usecase.BookUseCase, nc *nats.Conn) *BookHandler {
	return &BookHandler{
		usecase: u,
		nc:      nc,
	}
}

func (h *BookHandler) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.BookResponse, error) {
	if req == nil || req.Book == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	book := &domain.Book{
		Title:         req.Book.Title,
		Author:        req.Book.Author,
		Genre:         req.Book.Genre,
		Language:      req.Book.Language,
		Description:   req.Book.Description,
		Rating:        req.Book.Rating,
		Price:         req.Book.Price,
		Pages:         int(req.Book.Pages),
		PublishedDate: req.Book.PublishedDate,
	}

	created, err := h.usecase.CreateBook(ctx, book)
	if err != nil {
		return nil, err
	}

	evt := struct {
		Id     string `json:"id"`
		Title  string `json:"title"`
		Author string `json:"author"`
	}{
		Id:     created.ID.Hex(),
		Title:  created.Title,
		Author: created.Author,
	}
	if data, err := json.Marshal(evt); err == nil {
		h.nc.Publish("book.created", data)
	}

	return &pb.BookResponse{
		Book: &pb.Book{
			Id:            created.ID.Hex(),
			Title:         created.Title,
			Author:        created.Author,
			Genre:         created.Genre,
			Language:      created.Language,
			Description:   created.Description,
			Rating:        created.Rating,
			Price:         created.Price,
			Pages:         int32(created.Pages),
			PublishedDate: created.PublishedDate,
		},
	}, nil
}

func (h *BookHandler) GetBook(ctx context.Context, req *pb.BookID) (*pb.BookResponse, error) {
	book, err := h.usecase.GetBookByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.BookResponse{
		Book: &pb.Book{
			Id:            book.ID.Hex(),
			Title:         book.Title,
			Author:        book.Author,
			Genre:         book.Genre,
			Language:      book.Language,
			Description:   book.Description,
			Rating:        book.Rating,
			Price:         book.Price,
			Pages:         int32(book.Pages),
			PublishedDate: book.PublishedDate,
		},
	}, nil
}

func (h *BookHandler) ListAllBooks(ctx context.Context, _ *pb.Empty) (*pb.BookList, error) {
	books, err := h.usecase.ListBooks(ctx)
	if err != nil {
		return nil, err
	}

	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.BookResponse, error) {
	if req == nil || req.Book == nil {
		return nil, status.Error(codes.InvalidArgument, "empty update request")
	}
	objID, err := primitive.ObjectIDFromHex(req.Book.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid book ID")
	}
	book := &domain.Book{
		ID:            objID,
		Title:         req.Book.Title,
		Author:        req.Book.Author,
		Genre:         req.Book.Genre,
		Language:      req.Book.Language,
		Description:   req.Book.Description,
		Rating:        req.Book.Rating,
		Price:         req.Book.Price,
		Pages:         int(req.Book.Pages),
		PublishedDate: req.Book.PublishedDate,
	}
	updated, err := h.usecase.UpdateBook(ctx, book)
	if err != nil {
		return nil, err
	}
	return &pb.BookResponse{
		Book: &pb.Book{
			Id:            updated.ID.Hex(),
			Title:         updated.Title,
			Author:        updated.Author,
			Genre:         updated.Genre,
			Language:      updated.Language,
			Description:   updated.Description,
			Rating:        updated.Rating,
			Price:         updated.Price,
			Pages:         int32(updated.Pages),
			PublishedDate: updated.PublishedDate,
		},
	}, nil
}

func (h *BookHandler) DeleteBook(ctx context.Context, req *pb.BookID) (*pb.Empty, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "book ID is required")
	}
	if err := h.usecase.DeleteBook(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete book: %v", err)
	}
	return &pb.Empty{}, nil
}

func (h *BookHandler) ListBooksByGenre(ctx context.Context, req *pb.GenreRequest) (*pb.BookList, error) {
	books, err := h.usecase.ListBooksByGenre(ctx, req.Genre)
	if err != nil {
		return nil, err
	}
	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) ListBooksByAuthor(ctx context.Context, req *pb.AuthorRequest) (*pb.BookList, error) {
	books, err := h.usecase.ListBooksByAuthor(ctx, req.Author)
	if err != nil {
		return nil, err
	}
	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) ListBooksByLanguage(ctx context.Context, req *pb.LanguageRequest) (*pb.BookList, error) {
	books, err := h.usecase.ListBooksByLanguage(ctx, req.Language)
	if err != nil {
		return nil, err
	}
	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) ListTopRatedBooks(ctx context.Context, _ *pb.Empty) (*pb.BookList, error) {
	books, err := h.usecase.ListTopRated(ctx)
	if err != nil {
		return nil, err
	}
	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) ListNewArrivals(ctx context.Context, _ *pb.Empty) (*pb.BookList, error) {
	books, err := h.usecase.ListNewArrivals(ctx)
	if err != nil {
		return nil, err
	}
	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) SearchBooks(ctx context.Context, req *pb.SearchRequest) (*pb.BookList, error) {
	if req == nil || req.Keyword == "" {
		return nil, status.Error(codes.InvalidArgument, "keyword is required")
	}
	books, err := h.usecase.SearchBooks(ctx, req.Keyword)
	if err != nil {
		return nil, err
	}
	var res []*pb.Book
	for _, b := range books {
		res = append(res, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: res}, nil
}

func (h *BookHandler) RecommendBooks(ctx context.Context, req *pb.BookID) (*pb.BookList, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "book ID is required")
	}
	books, err := h.usecase.RecommendBooks(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	var pbBooks []*pb.Book
	for _, b := range books {
		pbBooks = append(pbBooks, &pb.Book{
			Id:            b.ID.Hex(),
			Title:         b.Title,
			Author:        b.Author,
			Genre:         b.Genre,
			Language:      b.Language,
			Description:   b.Description,
			Rating:        b.Rating,
			Price:         b.Price,
			Pages:         int32(b.Pages),
			PublishedDate: b.PublishedDate,
		})
	}
	return &pb.BookList{Books: pbBooks}, nil
}
