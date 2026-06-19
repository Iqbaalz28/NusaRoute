-- Deactivate the bogus "-02" hubs (seeded with coordinates clustered in the Java
-- Sea) so they're excluded from /hub/list and nearest-hub routing. The canonical
-- "-01" hubs (migration 001) have correct coordinates.
UPDATE hubs SET is_active = false WHERE code LIKE '%-02';
