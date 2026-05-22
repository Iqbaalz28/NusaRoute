import os

dc_path = 'e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/docker-compose.yml'

with open(dc_path, 'r', encoding='utf-8') as f:
    content = f.read()

content = content.replace('POSTGRES_PASSWORD: nusaroute_secret', 'POSTGRES_PASSWORD: ${DB_PASSWORD}')
content = content.replace('MONGO_INITDB_ROOT_PASSWORD: nusaroute_secret', 'MONGO_INITDB_ROOT_PASSWORD: ${MONGO_PASSWORD}')
content = content.replace('requirepass nusaroute_secret', 'requirepass ${REDIS_PASSWORD}')
content = content.replace('"nusaroute_secret"', '"${REDIS_PASSWORD}"')
content = content.replace('MINIO_ROOT_PASSWORD: nusaroute_secret', 'MINIO_ROOT_PASSWORD: ${MINIO_PASSWORD}')
content = content.replace('DB_PASSWORD=nusaroute_secret', 'DB_PASSWORD=${DB_PASSWORD}')
content = content.replace('REDIS_PASSWORD=nusaroute_secret', 'REDIS_PASSWORD=${REDIS_PASSWORD}')
content = content.replace('MONGO_URI=mongodb://nusaroute:nusaroute_secret@mongodb:27017', 'MONGO_URI=mongodb://nusaroute:${MONGO_PASSWORD}@mongodb:27017')
content = content.replace('MINIO_SECRET_KEY=nusaroute_secret', 'MINIO_SECRET_KEY=${MINIO_PASSWORD}')
content = content.replace('JWT_SECRET=nusaroute-jwt-secret-key-2026', 'JWT_SECRET=${JWT_SECRET}')

with open(dc_path, 'w', encoding='utf-8') as f:
    f.write(content)
print('updated docker-compose.yml')
