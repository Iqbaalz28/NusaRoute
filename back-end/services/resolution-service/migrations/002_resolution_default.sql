-- Make tickets.resolution non-null so SELECT * scans cleanly into the string
-- model field (fresh tickets had NULL resolution → "converting NULL to string").
UPDATE tickets SET resolution = '' WHERE resolution IS NULL;
ALTER TABLE tickets ALTER COLUMN resolution SET DEFAULT '';
