CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    cost INTEGER NOT NULL CHECK (cost >= 0),
    start_date CHAR(7) NOT NULL, 
    end_date CHAR(7), 
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_date_format CHECK (
        start_date ~ '^(0[1-9]|1[0-2])-\d{4}$' AND
        (end_date IS NULL OR end_date ~ '^(0[1-9]|1[0-2])-\d{4}$')
    )
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_service ON subscriptions(user_id, service_name);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
CREATE TRIGGER update_subscriptions_updated_at 
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE subscriptions IS 'User subscriptions to various services';
COMMENT ON COLUMN subscriptions.service_name IS 'Name of the service providing the subscription';
COMMENT ON COLUMN subscriptions.user_id IS 'UUID of the user who owns the subscription';
COMMENT ON COLUMN subscriptions.cost IS 'Monthly subscription cost in rubles (whole number)';
COMMENT ON COLUMN subscriptions.start_date IS 'Subscription start date in MM-YYYY format';
COMMENT ON COLUMN subscriptions.end_date IS 'Optional subscription end date in MM-YYYY format';