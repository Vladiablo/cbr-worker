CREATE TABLE IF NOT EXISTS exchange_rates (
    rate_date date NOT NULL,
    curr_code text NOT NULL CHECK (length(curr_code) = 3),
    curr_num_code int NOT NULL CHECK (curr_num_code >= 0 AND curr_num_code <= 999),
    rate decimal(20, 10) NOT NULL CHECK (rate > 0),

    PRIMARY KEY (rate_date, curr_code)
);
