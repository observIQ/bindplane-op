import { useEffect } from "react";
import { Navigate, useNavigate } from "react-router-dom";

export const RequireAuth: React.FC = ({ children }) => {
  const navigate = useNavigate();

  useEffect(() => {
    if (localStorage.getItem("user") == null) {
      navigate("/login");
    }
  }, [navigate]);

  useEffect(() => {
    function handleLoggedOut() {
      if (localStorage.getItem("user") == null) {
        navigate("/login");
      }
    }

    window.addEventListener("storage", handleLoggedOut);

    return () => document.removeEventListener("storage", handleLoggedOut);
  });

  useEffect(() => {
    async function verifyLogin() {
      const resp = await fetch("/verify");
      if (resp.status === 401) {
        localStorage.removeItem("user");
        navigate("/login");
      }
    }

    const timeout = setInterval(verifyLogin, 30 * 1000);

    return () => clearTimeout(timeout);
  }, [navigate]);

  return <>{children}</>;
};

export function withRequireLogin(FC: React.FC): React.FC {
  return () => (
    <RequireAuth>
      {localStorage.getItem("user") != null ? <FC /> : <Navigate to="/login" />}
    </RequireAuth>
  );
}
