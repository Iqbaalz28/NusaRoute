import os
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    service_name = 'unknown'
    m = re.search(r'services\\(.*?)\\cmd\\main.go', filepath)
    if m:
        service_name = m.group(1)
    else:
        m = re.search(r'api-gateway\\cmd\\main.go', filepath)
        if m:
            service_name = "api-gateway"

    if 'logger.InitLogger' not in content and 'func main()' in content:
        content = re.sub(r'func main\(\) \{', f'func main() {{\n\tlogger.InitLogger("{service_name}")', content)

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

for root, dirs, files in os.walk('e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/services'):
    for file in files:
        if file == 'main.go':
            process_file(os.path.join(root, file))

for root, dirs, files in os.walk('e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/api-gateway'):
    for file in files:
        if file == 'main.go':
            process_file(os.path.join(root, file))
