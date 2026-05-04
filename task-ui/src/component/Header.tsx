import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const LogoMark = () => (
  <svg width="38" height="38" viewBox="0 0 30 30" fill="none" aria-hidden="true">
    <rect width="30" height="30" rx="8" fill="url(#logo-gradient)" />
    <polyline points="8 15 13 20 22 10" stroke="#fff" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
    <defs>
      <linearGradient id="logo-gradient" x1="0" y1="0" x2="30" y2="30" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stopColor="#c084fc" />
        <stop offset="100%" stopColor="#7c3aed" />
      </linearGradient>
    </defs>
  </svg>
);

const UserIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
    <circle cx="12" cy="7" r="4" />
  </svg>
);

const SignOutIcon = () => (
  <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
    <polyline points="16 17 21 12 16 7" />
    <line x1="21" y1="12" x2="9" y2="12" />
  </svg>
);

export function Header() {
  const { token, userEmail, logout } = useAuth();
  const navigate = useNavigate();

  function handleSignOut() {
    logout();
    navigate("/login");
  }

  return (
    <header className="app-header">
      <div className="app-header__logo">
        <LogoMark />
        <span className="app-header__wordmark">Task Management Application</span>
      </div>
      {token ? (
        <div className="app-header__user">
          <div className="app-header__identity">
            <span className="app-header__avatar" aria-hidden="true">
              {userEmail?.charAt(0).toUpperCase() ?? "?"}
            </span>
            <span className="app-header__email">{userEmail}</span>
          </div>
          <button type="button" className="app-header__signout" onClick={handleSignOut} aria-label="Sign out" title="Sign out">
            <SignOutIcon />
          </button>
        </div>
      ) : (
        <button type="button" className="app-header__signin" onClick={() => navigate("/login")}>
          <UserIcon />
          Sign in
        </button>
      )}
    </header>
  );
}
