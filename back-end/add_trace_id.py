import os
import re

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    original_content = content

    # Add TraceID to BaseEvent
    # Example: BaseEvent: events.BaseEvent{EventID: uuid.New().String(), EventType: events.TopicCourierAssigned, Timestamp: time.Now(), Source: "dispatch-service"},
    # Or multiline
    
    # regex to find BaseEvent: events.BaseEvent{...}
    # and add TraceID: logger.GetTraceID(ctx), before the closing brace
    
    content = re.sub(r'(BaseEvent:\s*events\.BaseEvent\{[^}]*?)(})', r'\1, TraceID: logger.GetTraceID(ctx)\2', content)

    # Clean up multiple commas
    content = re.sub(r',\s*,\s*TraceID', ', TraceID', content)

    if content != original_content:
        # Check if we imported "github.com/nusaroute/pkg/logger"
        if '"github.com/nusaroute/pkg/logger"' not in content and 'logger.GetTraceID' in content:
            content = re.sub(r'import \(', 'import (\n\t"github.com/nusaroute/pkg/logger"', content, count=1)

        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Updated {filepath}")

for root, dirs, files in os.walk('e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/services'):
    for file in files:
        if file.endswith('.go'):
            process_file(os.path.join(root, file))
