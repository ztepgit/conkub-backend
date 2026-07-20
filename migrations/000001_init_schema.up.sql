-- migrations/000001_init_schema.up.sql

CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    description TEXT,
    venue VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    image_url TEXT,
    show_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE seats (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    row VARCHAR(10) NOT NULL,
    number INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(event_id, row, number)
);

CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    event_id INTEGER NOT NULL REFERENCES events(id),
    seat_id INTEGER NOT NULL REFERENCES seats(id),
    status VARCHAR(50) NOT NULL,
    stripe_session_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- สร้าง Index เพื่อให้ค้นหาที่นั่งของแต่ละคอนเสิร์ตได้เร็วขึ้น
CREATE INDEX idx_seats_event_id ON seats(event_id);