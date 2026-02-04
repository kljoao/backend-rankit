-- Criação da tabela rooms_history
CREATE TABLE IF NOT EXISTS rooms_history (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL,
    teacher_id TEXT NOT NULL,
    quiz_id TEXT NOT NULL,
    quiz_title_snapshot TEXT,
    status TEXT NOT NULL,
    total_questions INTEGER NOT NULL,
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (teacher_id) REFERENCES teachers (id),
    FOREIGN KEY (quiz_id) REFERENCES quizzes (id)
);

CREATE INDEX IF NOT EXISTS idx_rooms_history_teacher_id ON rooms_history (teacher_id);

CREATE INDEX IF NOT EXISTS idx_rooms_history_quiz_id ON rooms_history (quiz_id);

-- Criação da tabela room_players
CREATE TABLE IF NOT EXISTS room_players (
    id TEXT PRIMARY KEY,
    room_history_id TEXT NOT NULL,
    player_runtime_id TEXT, -- ID na sessão WebSocket
    nickname TEXT NOT NULL,
    score INTEGER NOT NULL,
    correct_count INTEGER NOT NULL,
    wrong_count INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (room_history_id) REFERENCES rooms_history (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_room_players_room_history_id ON room_players (room_history_id);

-- Criação da tabela room_questions (Agregados por pergunta)
CREATE TABLE IF NOT EXISTS room_questions (
    id TEXT PRIMARY KEY,
    room_history_id TEXT NOT NULL,
    question_index INTEGER NOT NULL,
    question_id TEXT, -- ID original da pergunta (se disponível)
    prompt_snapshot TEXT,
    correct_index INTEGER NOT NULL,
    count_a INTEGER DEFAULT 0,
    count_b INTEGER DEFAULT 0,
    count_c INTEGER DEFAULT 0,
    count_d INTEGER DEFAULT 0,
    correct_count INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (room_history_id) REFERENCES rooms_history (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_room_questions_room_history_id ON room_questions (room_history_id);

-- Criação da tabela room_answers (Respostas individuais - granularidade fina)
CREATE TABLE IF NOT EXISTS room_answers (
    id TEXT PRIMARY KEY,
    room_history_id TEXT NOT NULL,
    question_index INTEGER NOT NULL,
    room_player_id TEXT NOT NULL,
    selected_index INTEGER NOT NULL,
    is_correct BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (room_history_id) REFERENCES rooms_history (id) ON DELETE CASCADE,
    FOREIGN KEY (room_player_id) REFERENCES room_players (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_room_answers_room_history_id ON room_answers (room_history_id);

CREATE INDEX IF NOT EXISTS idx_room_answers_room_player_id ON room_answers (room_player_id);