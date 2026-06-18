"use client";

import React, { useState } from "react";
import { Navbar } from "../components/Navbar";
import { Footer } from "../components/Footer";
import { DashboardPage } from "../components/dashboard/DashboardPage";
import { TrackingPage } from "../components/TrackingPage";
import { OrdersPage } from "../components/OrdersPage";
import { ServicesPage } from "../components/ServicesPage";
import { CourierPage } from "../components/CourierPage";
import { HubsPage } from "../components/HubsPage";
import { AdminPage } from "../components/AdminPage";
import { LoginModal } from "../components/LoginModal";

export default function Home() {
  const [activePage, setActivePage] = useState("dashboard");
  const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);
  const [selectedAwb, setSelectedAwb] = useState("");

  const handleTrackPackage = (awb: string) => {
    setSelectedAwb(awb);
    setActivePage("tracking");
  };

  const renderPage = () => {
    switch (activePage) {
      case "dashboard":
        return (
          <DashboardPage
            onStartShipping={() => setActivePage("services")}
            onTrackPackage={handleTrackPackage}
            onNavigate={setActivePage}
          />
        );
      case "tracking":
        return <TrackingPage initialAwb={selectedAwb} />;
      case "orders":
        return <OrdersPage />;
      case "services":
        return <ServicesPage />;
      case "courier":
        return <CourierPage />;
      case "hubs":
        return <HubsPage />;
      case "admin":
        return <AdminPage />;
      default:
        return (
          <DashboardPage
            onStartShipping={() => setActivePage("services")}
            onTrackPackage={handleTrackPackage}
            onNavigate={setActivePage}
          />
        );
    }
  };

  return (
    <main style={{ background: 'var(--color-background)', minHeight: '100vh' }}>
      <Navbar
        activePage={activePage}
        setActivePage={setActivePage}
        onLoginClick={() => setIsLoginModalOpen(true)}
      />
      
      {renderPage()}

      <Footer />

      <LoginModal
        isOpen={isLoginModalOpen}
        onClose={() => setIsLoginModalOpen(false)}
      />
    </main>
  );
}
