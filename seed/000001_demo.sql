-- =================================================================================
-- ไฟล์: seed/000001_demo.sql
-- คำอธิบาย: ข้อมูลจำลอง (Mock Data) สำหรับตาราง events และ seats
-- =================================================================================

-- 1. เคลียร์ข้อมูลเก่าและรีเซ็ต ID กลับเป็น 1 (เพื่อให้รันสคริปต์นี้ซ้ำกี่ครั้งก็ได้)
TRUNCATE TABLE bookings, seats, events RESTART IDENTITY CASCADE;

-- 2. สร้างข้อมูล Events ทั้ง 6 งาน (ไม่ระบุ id ให้ PostgreSQL จัดการ Sequence อัตโนมัติ)
INSERT INTO events (
    name,
    artist,
    description,
    venue,
    category,
    image_url,
    show_time
) VALUES
('Summer Tour', 'The Sunshine Band', 'สัมผัสบรรยากาศดนตรีสดที่ดีที่สุดในหน้าร้อนนี้', 'Impact Arena', 'Pop', 'https://images.unsplash.com/photo-1498038432885-c6f3f1b912ee?auto=format&fit=crop&q=80&w=800', '2026-10-15 19:00:00+07'),
('Rock Effect', 'Thunder Strike', 'เตรียมตัวมามันส์ให้สุดเหวี่ยงไปกับคอนเสิร์ตร็อคสุดเดือด', 'Thunder Dome', 'Rock', 'https://images.unsplash.com/photo-1498038432885-c6f3f1b912ee?auto=format&fit=crop&q=80&w=800', '2026-11-05 20:00:00+07'),
('K-Pop Coming', 'Dream Girls', 'คอนเสิร์ต K-Pop ที่ทุกคนรอคอยแห่งปี', 'Rajamangala Stadium', 'K-Pop', 'https://images.unsplash.com/photo-1514525253161-7a46d19cd819?auto=format&fit=crop&q=80&w=800', '2026-12-10 18:00:00+07'),
('EDM Land', 'DJ Spark', 'แดนซ์กระจายไปกับบีทอิเล็กทรอนิกส์ที่ดีที่สุด', 'Bitec Bangna', 'EDM', 'https://images.unsplash.com/photo-1498038432885-c6f3f1b912ee?auto=format&fit=crop&q=80&w=800', '2026-09-25 21:00:00+07'),
('Hip-Hop Legends', 'MC Flow', 'การผสมผสานระหว่างฮิปฮอปยุคคลาสสิกและยุคใหม่', 'Live House', 'Hip-Hop', 'https://images.unsplash.com/photo-1498038432885-c6f3f1b912ee?auto=format&fit=crop&q=80&w=800', '2026-08-30 19:30:00+07'),
('Symphony Night', 'Grand Orchestra', 'ค่ำคืนสุดผ่อนคลายกับดนตรีคลาสสิก', 'Thailand Cultural Centre', 'Classical', 'https://images.unsplash.com/photo-1498038432885-c6f3f1b912ee?auto=format&fit=crop&q=80&w=800', '2026-08-15 19:00:00+07');

-- 3. สร้างข้อมูลที่นั่ง 90 ที่นั่งต่อ 1 Event (A-I แถวละ 10 ที่นั่ง)
-- โดยกำหนดประเภท (VIP/REGULAR) และ ราคา ตามแต่ละ Event
INSERT INTO seats (
    event_id,
    row,
    number,
    seat_type,
    price,
    status
)
SELECT
    e.id,
    r.row_name,
    n.num,

    -- กำหนดแถว A, B, C เป็น VIP และ D ถึง I เป็น REGULAR
    CASE
        WHEN r.row_name IN ('A', 'B', 'C') THEN 'VIP'
        ELSE 'REGULAR'
    END AS seat_type,

    -- กำหนดราคาตาม Event และปรับ VIP ให้แพงกว่า Regular 30%
    CASE
        WHEN e.name = 'Summer Tour' THEN
            CASE
                WHEN r.row_name IN ('A', 'B', 'C') THEN 3250.00
                ELSE 2500.00
            END

        WHEN e.name = 'Rock Effect' THEN
            CASE
                WHEN r.row_name IN ('A', 'B', 'C') THEN 3900.00
                ELSE 3000.00
            END

        WHEN e.name = 'K-Pop Coming' THEN
            CASE
                WHEN r.row_name IN ('A', 'B', 'C') THEN 5850.00
                ELSE 4500.00
            END

        WHEN e.name = 'EDM Land' THEN
            CASE
                WHEN r.row_name IN ('A', 'B', 'C') THEN 2600.00
                ELSE 2000.00
            END

        WHEN e.name = 'Hip-Hop Legends' THEN
            CASE
                WHEN r.row_name IN ('A', 'B', 'C') THEN 1950.00
                ELSE 1500.00
            END

        WHEN e.name = 'Symphony Night' THEN
            CASE
                WHEN r.row_name IN ('A', 'B', 'C') THEN 4550.00
                ELSE 3500.00
            END
    END AS price,

    'AVAILABLE' AS status

FROM events e
CROSS JOIN (
    VALUES
        ('A'),
        ('B'),
        ('C'),
        ('D'),
        ('E'),
        ('F'),
        ('G'),
        ('H'),
        ('I')
) AS r(row_name)
CROSS JOIN generate_series(1, 10) AS n(num);