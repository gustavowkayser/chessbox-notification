CREATE TABLE IF NOT EXISTS notifications (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	device_push_token VARCHAR(255) NOT NULL,
	title VARCHAR(255) NOT NULL,
	message VARCHAR(255) NOT NULL,
	data JSONB,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	tournament_id UUID REFERENCES tournaments(id)
);
