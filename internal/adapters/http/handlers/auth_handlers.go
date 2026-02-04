package handlers

import (
	"encoding/json"
	"net/http"
	"rankit/internal/adapters/http/middlewares"
	"rankit/internal/application/usecases"
)

// AuthHandler agrupa os handlers de autenticação.
type AuthHandler struct {
	registerUC *usecases.RegisterTeacherUseCase
	loginUC    *usecases.LoginTeacherUseCase
	getMeUC    *usecases.GetMeUseCase
}

// NewAuthHandler cria um novo handler de autenticação.
func NewAuthHandler(
	registerUC *usecases.RegisterTeacherUseCase,
	loginUC *usecases.LoginTeacherUseCase,
	getMeUC *usecases.GetMeUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		getMeUC:    getMeUC,
	}
}

// Register godoc
// @Summary Cadastra um novo professor
// @Description Cria uma conta de professor com nome, email e senha.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body usecases.RegisterInput true "Dados de cadastro"
// @Success 201 {object} usecases.RegisterOutput
// @Failure 400 {object} map[string]string "Erro de validação"
// @Failure 409 {object} map[string]string "Email já cadastrado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input usecases.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	output, err := h.registerUC.Execute(r.Context(), input)
	if err != nil {
		if err == usecases.ErrEmailDuplicado {
			http.Error(w, err.Error(), http.StatusConflict) // 409
			return
		}
		if err.Error() == "o nome é obrigatório" || err.Error() == "o email é inválido" || err.Error() == "a senha deve ter no mínimo 6 caracteres" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(output)
}

// Login godoc
// @Summary Autentica um professor
// @Description Realiza login com email e senha e retorna um token JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body usecases.LoginInput true "Credenciais"
// @Success 200 {object} usecases.LoginOutput
// @Failure 401 {object} map[string]string "Credenciais inválidas"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input usecases.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	output, err := h.loginUC.Execute(r.Context(), input)
	if err != nil {
		if err == usecases.ErrCredenciaisInvalidas {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

// GetMe godoc
// @Summary Retorna dados do professor logado
// @Description Obtém detalhes do perfil do usuário autenticado via token JWT.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} teacher.Teacher
// @Failure 401 {object} map[string]string "Não autenticado"
// @Failure 500 {object} map[string]string "Erro interno"
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	output, err := h.getMeUC.Execute(r.Context(), userID)
	if err != nil {
		http.Error(w, "Erro ao buscar perfil", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}
