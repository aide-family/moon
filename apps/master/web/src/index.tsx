import React from "react";
import ReactDOM from "react-dom/client";
import App from "./pages/App";
import "./public-path";
import "@/style/global.less";
import "./index.css";

const root = ReactDOM.createRoot(
  document.getElementById("dashboard-app") as HTMLElement
);

window.addEventListener("unmount", function () {
  console.log("unmount");
  root.unmount();
});

root.render(<App />);
