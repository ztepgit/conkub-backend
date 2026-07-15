CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    venue VARCHAR(255) NOT NULL,
    show_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE seats (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    row VARCHAR(10) NOT NULL,
    number INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'AVAILABLE' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    event_id INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_id INT NOT NULL UNIQUE REFERENCES seats(id) ON DELETE CASCADE, -- กัน Double Booking ชั้นแรกที่ DB
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- สร้าง Index เพื่อให้ค้นหาที่นั่งของแต่ละคอนเสิร์ตได้เร็วขึ้น
CREATE INDEX idx_seats_event_id ON seats(event_id);