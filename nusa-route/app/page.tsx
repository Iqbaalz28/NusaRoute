"use client";

import React, { useState } from "react";
import { Navbar } from "../components/Navbar";
import { Footer } from "../components/Footer";
import { DashboardPage } from "../components/dashboard/DashboardPage";
import { TrackingPage } from "../components/TrackingPage";
import { OrdersPage } from "../components/OrdersPage";
import { ServicesPage } from "../components/ServicesPage";
import { HubsPage } from "../components/HubsPage";
import { LoginModal } from "../components/LoginModal";

export default function Home() {
  const [activePage, setActivePage] = useState("dashboard");
  const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);

  const renderPage = () => {
    switch (activePage) {
      case "dashboard":
        return (
          <DashboardPage
            onStartShipping={() => setActivePage("services")}
            onTrackPackage={() => setActivePage("tracking")}
          />
        );
      case "tracking":
        return <TrackingPage />;
      case "orders":
        return <OrdersPage />;
      case "services":
        return <ServicesPage />;
      case "hubs":
        return <HubsPage />;
      default:
        return (
          <DashboardPage
            onStartShipping={() => setActivePage("services")}
            onTrackPackage={() => setActivePage("tracking")}
          />
        );
    }
  };

  return (
    <main className="main">
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
