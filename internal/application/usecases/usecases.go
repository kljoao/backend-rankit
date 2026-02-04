package usecases

import (
	"context"
	"errors"
	"rankit/internal/domain/teacher"
	"rankit/internal/ports"
)

// Casos de erro comuns
var (
	ErrEmailDuplicado       = errors.New("email já cadastrado")
	ErrCredenciaisInvalidas = errors.New("email ou senha inválidos")
	ErrUsuarioNaoEncontrado = errors.New("usuário não encontrado")
)

// RegisterTeacherUseCase coordena o registro de um novo professor.
type RegisterTeacherUseCase struct {
	repo   ports.TeacherRepository
	hasher ports.PasswordHasher
}

func NewRegisterTeacherUseCase(repo ports.TeacherRepository, hasher ports.PasswordHasher) *RegisterTeacherUseCase {
	return &RegisterTeacherUseCase{repo: repo, hasher: hasher}
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type RegisterOutput struct {
	ID    string
	Name  string
	Email string
}

func (uc *RegisterTeacherUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// 1. Verifica se email já existe
	existing, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailDuplicado
	}

	// 2. Cria entidade Teacher com validações de domínio
	newTeacher, err := teacher.NewTeacher(input.Name, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	// 3. Hash da senha
	hashedPassword, err := uc.hasher.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	newTeacher.SetPassword(hashedPassword)

	// 4. Persiste
	if err := uc.repo.Create(ctx, newTeacher); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		ID:    newTeacher.ID,
		Name:  newTeacher.Name,
		Email: newTeacher.Email,
	}, nil
}

// LoginTeacherUseCase coordena o login.
type LoginTeacherUseCase struct {
	repo         ports.TeacherRepository
	hasher       ports.PasswordHasher
	tokenService ports.TokenService
}

func NewLoginTeacherUseCase(repo ports.TeacherRepository, hasher ports.PasswordHasher, tokenService ports.TokenService) *LoginTeacherUseCase {
	return &LoginTeacherUseCase{repo: repo, hasher: hasher, tokenService: tokenService}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AccessToken string
	ExpiresIn   int64 // Segundos
}

func (uc *LoginTeacherUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// 1. Busca usuário
	t, err := uc.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrCredenciaisInvalidas
	}

	// 2. Valida senha
	if err := uc.hasher.ComparePassword(t.PasswordHash, input.Password); err != nil {
		return nil, ErrCredenciaisInvalidas
	}

	// 3. Gera Token
	token, expiresIn, err := uc.tokenService.GenerateToken(t.ID)
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken: token,
		ExpiresIn:   expiresIn,
	}, nil
}

// GetMeUseCase retorna dados do usuário logado.
type GetMeUseCase struct {
	repo ports.TeacherRepository
}

func NewGetMeUseCase(repo ports.TeacherRepository) *GetMeUseCase {
	return &GetMeUseCase{repo: repo}
}

func (uc *GetMeUseCase) Execute(ctx context.Context, userID string) (*teacher.Teacher, error) {
	t, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrUsuarioNaoEncontrado
	}
	return t, nil
}
