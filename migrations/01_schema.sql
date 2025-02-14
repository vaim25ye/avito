CREATE TABLE IF NOT EXISTS "user" (
                                      user_id   SERIAL PRIMARY KEY,
                                      name      VARCHAR(255) NOT NULL,
    password  VARCHAR(255) NOT NULL,
    balance   INT NOT NULL DEFAULT 0
    );

CREATE TABLE IF NOT EXISTS merch (
                                     merch_id SERIAL PRIMARY KEY,
                                     type     VARCHAR(255) NOT NULL,
    price    INT NOT NULL
    );

CREATE TABLE IF NOT EXISTS purchase (
                                        purchase_id SERIAL PRIMARY KEY,
                                        user_id     INT NOT NULL,
                                        merch_id    INT NOT NULL,
                                        amount      INT NOT NULL,
                                        FOREIGN KEY (user_id)  REFERENCES "user" (user_id),
    FOREIGN KEY (merch_id) REFERENCES merch (merch_id)
    );

CREATE TABLE IF NOT EXISTS operation (
                                         operation_id SERIAL PRIMARY KEY,
                                         fromUser     INT NOT NULL,
                                         toUser       INT NOT NULL,
                                         amount       INT NOT NULL,
                                         FOREIGN KEY (fromUser) REFERENCES "user" (user_id),
    FOREIGN KEY (toUser)   REFERENCES "user" (user_id)
    );