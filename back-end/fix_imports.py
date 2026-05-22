import os
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    # If it's a generated functional test and we don't use context, repository, service, remove them.
    # Simple way: just remove the unused imports since the boilerplate doesn't use them yet.
    if 'func Test' in content and 'client.Ping' in content or 'db.Ping' in content:
        content = re.sub(r'\t"context"\n', '', content)
        content = re.sub(r'\t"github.com/nusaroute/services/.*-service/internal/repository"\n', '', content)
        content = re.sub(r'\t"github.com/nusaroute/services/.*-service/internal/service"\n', '', content)

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Updated {filepath}")

for root, dirs, files in os.walk('e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/services'):
    for file in files:
        if file.endswith('_functional_test.go') and file not in ['order_functional_test.go', 'user_functional_test.go']:
            process_file(os.path.join(root, file))
