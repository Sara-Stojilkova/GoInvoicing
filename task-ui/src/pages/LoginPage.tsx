import { useState } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { ApiError } from "../api/error";

const WatermarkMark = () => (
  <svg className="auth-watermark" viewBox="0 0 120 120" fill="none" aria-hidden="true">
    <rect width="120" height="120" rx="28" fill="url(#wm-grad)" />
    <polyline points="28 60 50 82 92 38" stroke="#fff" strokeWidth="9" strokeLinecap="round" strokeLinejoin="round" />
    <defs>
      <linearGradient id="wm-grad" x1="0" y1="0" x2="120" y2="120" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stopColor="#c084fc" stopOpacity="0.3" />
        <stop offset="100%" stopColor="#7c3aed" stopOpacity="0.15" />
      </linearGradient>
    </defs>
  </svg>
);

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const registered = searchParams.get("registered") === "true";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await login(email, password);
      navigate("/");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError("Invalid email or password.");
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="auth-page">
      <div className="auth-panel auth-panel--left" aria-hidden="true">
        <WatermarkMark />
        <div className="auth-panel__copy">
          <p className="auth-tagline">Coordinate work.<br />Ship faster.</p>
        </div>
      </div>

      <div className="auth-panel auth-panel--right">
        <div className="auth-form-wrap">
          <h1 className="auth-heading">Welcome back</h1>
          <p className="auth-subheading">Sign in to your workspace</p>

          {registered && (
            <div className="auth-alert auth-alert--success">
              Account created — please sign in.
            </div>
          )}
          {error && (
            <div className="auth-alert auth-alert--error">{error}</div>
          )}

          <form className="auth-form" onSubmit={handleSubmit} noValidate>
            <div className="auth-field">
              <label className="auth-label" htmlFor="login-email">Email</label>
              <input
                id="login-email"
                className="auth-input"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                autoFocus
              />
            </div>
            <div className="auth-field">
              <label className="auth-label" htmlFor="login-password">Password</label>
              <input
                id="login-password"
                className="auth-input"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </div>
            <button className="auth-submit" type="submit" disabled={loading}>
              {loading ? "Signing in…" : "Sign in"}
            </button>
          </form>

          <p className="auth-footer">
            No account?{" "}
            <Link to="/register" className="auth-link">Create one</Link>
          </p>
        </div>
      </div>
    </main>
  );
}
