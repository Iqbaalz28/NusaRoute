-- Take-job model: assignments start as OPEN with no courier, then a courier
-- claims one (first-come-first-served). courier_id must therefore be nullable.
ALTER TABLE assignments ALTER COLUMN courier_id DROP NOT NULL;

-- Status now also includes 'OPEN' (job available to claim). No enum change
-- needed since status is VARCHAR; documented here for clarity:
--   OPEN -> ASSIGNED (claimed) -> PICKED_UP -> COMPLETED
