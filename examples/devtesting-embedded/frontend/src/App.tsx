import { useState } from "react";
import {
  captureException,
  getDistributedTraceId,
} from "@tracewayapp/frontend";

const API_BASE = "http://localhost:8080";

export default function App() {
  const [log, setLog] = useState<string[]>([]);

  function addLog(msg: string) {
    setLog((prev) => [...prev, `[${new Date().toLocaleTimeString()}] ${msg}`]);
  }

  async function callErrorEndpoint() {
    addLog("Calling GET /api/test-error ...");
    try {
      const response = await fetch(`${API_BASE}/api/test-error`);
      const traceId = getDistributedTraceId(response);
      addLog(`Response: ${response.status} — traceway-trace-id: ${traceId ?? "not present"}`);

      if (!response.ok) {
        const body = await response.json();
        const error = new Error(body.error || "Backend error");
        captureException(error, { distributedTraceId: traceId });
        addLog(`Captured frontend exception with distributedTraceId=${traceId}`);
      }
    } catch (err: any) {
      addLog(`Network error: ${err.message}`);
      captureException(err);
    }
  }

  async function callSuccessEndpoint() {
    addLog("Calling GET /api/test-success ...");
    try {
      const response = await fetch(`${API_BASE}/api/test-success`);
      const traceId = getDistributedTraceId(response);
      addLog(`Response: ${response.status} — traceway-trace-id: ${traceId ?? "not present"}`);
    } catch (err: any) {
      addLog(`Network error: ${err.message}`);
    }
  }

  return (
    <div style={{ fontFamily: "system-ui", maxWidth: 700, margin: "40px auto", padding: "0 20px" }}>
      <h1>Distributed Trace Test</h1>
      <p>
        Click the error button to trigger a backend 500 error. The response's{" "}
        <code>traceway-trace-id</code> header is captured and attached to the frontend exception,
        linking both projects in the dashboard.
      </p>

      <div style={{ display: "flex", gap: 12, marginBottom: 24 }}>
        <button onClick={callErrorEndpoint} style={btnStyle("#dc2626")}>
          Call Backend Error Endpoint
        </button>
        <button onClick={callSuccessEndpoint} style={btnStyle("#16a34a")}>
          Call Backend Success Endpoint
        </button>
        <button onClick={() => setLog([])} style={btnStyle("#6b7280")}>
          Clear Log
        </button>
      </div>

      <div
        style={{
          background: "#111",
          color: "#0f0",
          padding: 16,
          borderRadius: 8,
          fontFamily: "monospace",
          fontSize: 13,
          minHeight: 200,
          whiteSpace: "pre-wrap",
        }}
      >
        {log.length === 0 ? "Waiting for actions..." : log.join("\n")}
      </div>
    </div>
  );
}

function btnStyle(bg: string): React.CSSProperties {
  return {
    background: bg,
    color: "#fff",
    border: "none",
    padding: "10px 18px",
    borderRadius: 6,
    cursor: "pointer",
    fontSize: 14,
    fontWeight: 600,
  };
}
