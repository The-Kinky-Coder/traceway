import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { TracewayProvider } from "@tracewayapp/react";
import App from "./App";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <TracewayProvider connectionString="frontend-dev-token@http://localhost:8082/api/report">
      <App />
    </TracewayProvider>
  </StrictMode>,
);
