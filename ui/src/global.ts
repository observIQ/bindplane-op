import { useNavigate } from "react-router-dom";

declare global {
  interface Window {
    navigate: ReturnType<typeof useNavigate> | null;
  }
}
