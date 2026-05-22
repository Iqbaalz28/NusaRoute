import os, glob

back_end_dir = 'e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end'

for p in glob.glob(os.path.join(back_end_dir, '**', 'cmd', 'main.go'), recursive=True):
    with open(p, 'r', encoding='utf-8') as f:
        content = f.read()
    
    content = content.replace('"nusaroute_secret"', '""')
    content = content.replace('"mongodb://nusaroute:nusaroute_secret@localhost:27017"', '"mongodb://localhost:27017"')
    content = content.replace('"nusaroute-jwt-secret-key-2026"', '""')

    with open(p, 'w', encoding='utf-8') as f:
        f.write(content)

print("done")
