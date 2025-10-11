ALTER TABLE quizzes
ADD COLUMN study_group_id UUID REFERENCES study_group(id) ON DELETE SET NULL;
