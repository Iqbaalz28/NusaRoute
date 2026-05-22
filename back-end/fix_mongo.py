import os

def fix_mongo_test(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    content = content.replace('Database: dbname', 'DBName: dbname')
    content = content.replace('client.Disconnect(', 'client.Client().Disconnect(')
    content = content.replace('client.Ping(', 'client.Client().Ping(')

    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Fixed {filepath}")

base = 'e:/UNIVERSITAS PENDIDIKAN INDONESIA/SEMESTER 4/DISTRIBUTED PARALLEL AND CLOUD COMPUTING/Tubes/NusaRoute/back-end/services'
fix_mongo_test(os.path.join(base, 'tracking-service/internal/service/tracking_functional_test.go'))
fix_mongo_test(os.path.join(base, 'notification-service/internal/service/notification_functional_test.go'))
