import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { ApiError } from "../api/error";

type AgencyMode = "create" | "join";

const WatermarkMark = () => (
  <svg className="auth-watermark" viewBox="0 0 120 120" fill="none" aria-hidden="true">
    <rect width="120" height="120" rx="28" fill="url(#wm-grad-reg)" />
    <polyline points="28 60 50 82 92 38" stroke="#fff" strokeWidth="9" strokeLinecap="round" strokeLinejoin="round" />
    <defs>
      <linearGradient id="wm-grad-reg" x1="0" y1="0" x2="120" y2="120" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stopColor="#c084fc" stopOpacity="0.3" />
        <stop offset="100%" stopColor="#7c3aed" stopOpacity="0.15" />
      </linearGradient>
    </defs>
  </svg>
);

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [agencyMode, setAgencyMode] = useState<AgencyMode>("create");
  const [agencyName, setAgencyName] = useState("");
  const [agencyId, setAgencyId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await register({
        full_name: fullName,
        email,
        password,
        ...(agencyMode === "create" ? { agency_name: agencyName } : { agency_id: agencyId }),
      });
      navigate("/login?registered=true");
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 409) setError("Email already registered.");
        else if (err.status === 404) setError("Agency not found.");
        else setError("Something went wrong. Please try again.");
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
          <p className="auth-tagline">Your agency.<br />Your workflow.</p>
        </div>
      </div>

      <div className="auth-panel auth-panel--right">
        <div className="auth-form-wrap">
          <h1 className="auth-heading">Create account</h1>
          <p className="auth-subheading">Get your team up and running</p>

          {error && (
            <div className="auth-alert auth-alert--error">{error}</div>
          )}

          <form className="auth-form" onSubmit={handleSubmit} noValidate>
            <div className="auth-field">
              <label className="auth-label" htmlFor="reg-name">Full name</label>
              <input
                id="reg-name"
                className="auth-input"
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
                autoComplete="name"
                autoFocus
              />
            </div>
            <div className="auth-field">
              <label className="auth-label" htmlFor="reg-email">Email</label>
              <input
                id="reg-email"
                className="auth-input"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
              />
            </div>
            <div className="auth-field">
              <label className="auth-label" htmlFor="reg-password">Password</label>
              <input
                id="reg-password"
                className="auth-input"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete="new-password"
              />
            </div>

            <div className="auth-field">
              <span className="auth-label">Agency</span>
              <div className="auth-segment" role="group" aria-label="Agency option">
                <button
                  type="button"
                  className={`auth-segment__btn${agencyMode === "create" ? " auth-segment__btn--active" : ""}`}
                  onClick={() => setAgencyMode("create")}
                >
                  Create new
                </button>
                <button
                  type="button"
                  className={`auth-segment__btn${agencyMode === "join" ? " auth-segment__btn--active" : ""}`}
                  onClick={() => setAgencyMode("join")}
                >
                  Join existing
                </button>
              </div>
              {agencyMode === "create" ? (
                <input
                  className="auth-input"
                  type="text"
                  value={agencyName}
                  onChange={(e) => setAgencyName(e.target.value)}
                  required
                  placeholder="Agency name"
                  aria-label="Agency name"
                />
              ) : (
                <input
                  className="auth-input"
                  type="text"
                  value={agencyId}
                  onChange={(e) => setAgencyId(e.target.value)}
                  required
                  placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                  aria-label="Agency ID"
                />
              )}
            </div>

            <button className="auth-submit" type="submit" disabled={loading}>
              {loading ? "Creating account…" : "Create account"}
            </button>
          </form>

          <p className="auth-footer">
            Already have an account?{" "}
            <Link to="/login" className="auth-link">Sign in</Link>
          </p>
        </div>
      </div>
    </main>
  );
}
