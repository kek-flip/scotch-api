ALTER TABLE likes
    ALTER user_id SET NOT NULL,
    ALTER liked_user SET NOT NULL;