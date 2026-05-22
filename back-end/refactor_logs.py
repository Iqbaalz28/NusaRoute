import os
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    original_content = content

    # Add imports
    if '"context"' not in content and 'log.' in content:
        content = re.sub(r'import \(', 'import (\n\t"context"\n\t"fmt"\n\t"github.com/nusaroute/pkg/logger"', content, count=1)
    elif '"fmt"' not in content and 'log.' in content:
        content = re.sub(r'import \(', 'import (\n\t"fmt"\n\t"github.com/nusaroute/pkg/logger"', content, count=1)
    elif '"github.com/nusaroute/pkg/logger"' not in content and 'log.' in content:
        content = re.sub(r'import \(', 'import (\n\t"github.com/nusaroute/pkg/logger"', content, count=1)

    # Replace log.Printf(..., args) with logger.Info(context.Background(), fmt.Sprintf(..., args))
    # This is a basic regex, handles single line log.Printf
    content = re.sub(r'log\.Printf\((.*?)\)', r'logger.Info(context.Background(), fmt.Sprintf(\1))', content)
    
    # Replace log.Println(...) with logger.Info(context.Background(), fmt.Sprint(...))
    content = re.sub(r'log\.Println\((.*?)\)', r'logger.Info(context.Background(), fmt.Sprint(\1))', content)

    # Replace log.Fatal(...) with logger.Log.Fatal(...)
    content = re.sub(r'log\.Fatal\((.*?)\)', r'logger.Log.Fatal(fmt.Sprint(\1))', content)
    content = re.sub(r'log\.Fatalf\((.*?)\)', r'logger.Log.Fatal(fmt.Sprintf(\1))', content)

    if content != original_content:
        # Check if we imported "log" and remove it if it's no longer used
        if 'log.' not in content:
             content = re.sub(r'\t"log"\n', '', content)

        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Updated {filepath}")

for root, dirs, files in os.walk('e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/services'):
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go'):
            process_file(os.path.join(root, file))

for root, dirs, files in os.walk('e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/api-gateway'):
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go'):
            process_file(os.path.join(root, file))
