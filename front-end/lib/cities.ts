// Supported Indonesian cities with real coordinates. A fixed list keeps the tariff
// calculator from receiving invalid free-text input and lets the courier console
// simulate "being" in a city without a real GPS fix — both reuse this single source.
export type City = { name: string; lat: number; lng: number };

export const CITIES: City[] = [
  { name: "Jakarta", lat: -6.2088, lng: 106.8456 },
  { name: "Bogor", lat: -6.5950, lng: 106.8166 },
  { name: "Bandung", lat: -6.9175, lng: 107.6191 },
  { name: "Semarang", lat: -6.9667, lng: 110.4167 },
  { name: "Yogyakarta", lat: -7.7956, lng: 110.3695 },
  { name: "Surabaya", lat: -7.2575, lng: 112.7521 },
  { name: "Malang", lat: -7.9819, lng: 112.6265 },
  { name: "Denpasar (Bali)", lat: -8.6705, lng: 115.2126 },
  { name: "Medan", lat: 3.5952, lng: 98.6722 },
  { name: "Padang", lat: -0.9471, lng: 100.4172 },
  { name: "Pekanbaru", lat: 0.5071, lng: 101.4478 },
  { name: "Palembang", lat: -2.9761, lng: 104.7754 },
  { name: "Batam", lat: 1.0456, lng: 104.0305 },
  { name: "Pontianak", lat: -0.0263, lng: 109.3425 },
  { name: "Banjarmasin", lat: -3.3194, lng: 114.5908 },
  { name: "Balikpapan", lat: -1.2379, lng: 116.8529 },
  { name: "Makassar", lat: -5.1477, lng: 119.4327 },
  { name: "Manado", lat: 1.4748, lng: 124.8421 },
];
