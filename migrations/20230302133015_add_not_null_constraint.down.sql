ALTER TABLE likes
    ALTER user_id DROP NOT NULL,
    ALTER liked_user DROP NOT NULL;