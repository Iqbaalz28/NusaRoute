const fs = require('fs');
const { execSync } = require('child_process');

const API_URL = 'http://localhost:8080/api/v1';

async function seedData() {
  console.log('🌱 Memulai proses seeding data NusaRoute...');
  console.log('⚠️ Pastikan seluruh microservices dan API Gateway sedang berjalan di port 8080.');

  try {
    // 1. Seed User
    console.log('\n[1/4] Membuat akun pengguna...');
    const userRes = await fetch(`${API_URL}/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: 'demo@nusaroute.com',
        password: 'password123',
        first_name: 'Budi',
        last_name: 'Santoso'
      })
    });
    
    let token = '';
    
    if (userRes.ok) {
      const data = await userRes.json();
      console.log('✅ User berhasil dibuat: demo@nusaroute.com');
      // Login to get token
      const loginRes = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'demo@nusaroute.com',
          password: 'password123'
        })
      });
      const loginData = await loginRes.json();
      token = loginData.data.token;
    } else if (userRes.status === 409) {
      console.log('ℹ️ User sudah ada. Melakukan login...');
      const loginRes = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'demo@nusaroute.com',
          password: 'password123'
        })
      });
      if (loginRes.ok) {
        const loginData = await loginRes.json();
        token = loginData.data.token;
      } else {
        throw new Error('Gagal login user demo');
      }
    } else {
      console.log(await userRes.text());
      throw new Error(`Gagal membuat user: ${userRes.status}`);
    }

    console.log('🔑 Token otentikasi didapatkan.');

    // 2. Seed Orders
    console.log('\n[2/4] Membuat data pesanan (Orders)...');
    
    const ordersToCreate = [
      {
        sender_name: 'Budi Santoso',
        sender_phone: '08123456789',
        sender_address: 'Jl. Merdeka No. 10, Bandung, Jawa Barat',
        sender_lat: -6.917464,
        sender_lng: 107.619123,
        receiver_name: 'Siti Aminah',
        receiver_phone: '08987654321',
        receiver_address: 'Jl. Sudirman No. 25, Jakarta Selatan, DKI Jakarta',
        receiver_lat: -6.225014,
        receiver_lng: 106.800316,
        item_description: 'Dokumen Penting',
        weight_kg: 1.5,
        length_cm: 30,
        width_cm: 20,
        height_cm: 5,
        service_type: 'EXPRESS',
        is_insured: false,
        insured_value: 0,
        shipping_cost: 35000,
        insurance_cost: 0,
        total_cost: 35000
      },
      {
        sender_name: 'Budi Santoso',
        sender_phone: '08123456789',
        sender_address: 'Jl. Merdeka No. 10, Bandung, Jawa Barat',
        sender_lat: -6.917464,
        sender_lng: 107.619123,
        receiver_name: 'Ahmad Yani',
        receiver_phone: '08555555555',
        receiver_address: 'Jl. Pahlawan No. 5, Surabaya, Jawa Timur',
        receiver_lat: -7.250445,
        receiver_lng: 112.768845,
        item_description: 'Pakaian dan Sepatu',
        weight_kg: 3.0,
        length_cm: 40,
        width_cm: 30,
        height_cm: 20,
        service_type: 'REGULAR',
        is_insured: true,
        insured_value: 500000,
        shipping_cost: 55000,
        insurance_cost: 2500,
        total_cost: 57500
      },
      {
        sender_name: 'Budi Santoso',
        sender_phone: '08123456789',
        sender_address: 'Jl. Merdeka No. 10, Bandung, Jawa Barat',
        sender_lat: -6.917464,
        sender_lng: 107.619123,
        receiver_name: 'Toko Elektronik Makmur',
        receiver_phone: '08111111111',
        receiver_address: 'Jl. Gajah Mada No. 100, Denpasar, Bali',
        receiver_lat: -8.650000,
        receiver_lng: 115.216667,
        item_description: 'Komponen Komputer',
        weight_kg: 15.0,
        length_cm: 50,
        width_cm: 40,
        height_cm: 30,
        service_type: 'CARGO',
        is_insured: true,
        insured_value: 2000000,
        shipping_cost: 150000,
        insurance_cost: 10000,
        total_cost: 160000
      }
    ];

    const createdOrders = [];
    
    for (const order of ordersToCreate) {
      const orderRes = await fetch(`${API_URL}/orders`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify(order)
      });
      
      if (orderRes.ok) {
        const orderData = await orderRes.json();
        console.log(`✅ Pesanan berhasil dibuat: ${orderData.data.awb}`);
        createdOrders.push(orderData.data);
      } else {
        console.error(`❌ Gagal membuat pesanan: ${orderRes.status}`);
        const err = await orderRes.text();
        console.error(err);
      }
    }

    // 3. Trigger Payments for Orders
    console.log('\n[3/5] Melakukan simulasi pembayaran...');
    for (const order of createdOrders) {
      console.log(`ℹ️ Pesanan ${order.awb} berada dalam status ${order.status}`);
    }

    // 4. Bulk Seeding via Go Script
    console.log('\n[4/5] Melakukan seeding data agregasi ke Database (Go Script)...');
    try {
      execSync('go run ../back-end/scripts/seed.go', { stdio: 'inherit' });
      console.log('✅ Bulk seeding Database selesai.');
    } catch (e) {
      console.error('❌ Gagal menjalankan script Go bulk seeding. Pastikan Go terinstall dan database menyala.', e.message);
    }

    console.log('\n[5/5] Selesai!');
    console.log('✅ Proses seeding data berhasil. Cek Dashboard untuk melihat Real Data (Orders, Stats, Volume).');
    
  } catch (error) {
    if (error.cause && error.cause.code === 'ECONNREFUSED') {
      console.error('\n❌ GAGAL: Tidak dapat terhubung ke API Gateway.');
      console.error('💡 Solusi: Pastikan backend microservices dan API Gateway sedang menyala (misal di port 8080).');
    } else {
      console.error('\n❌ Terjadi kesalahan saat seeding data:', error.message);
    }
  }
}

seedData();
