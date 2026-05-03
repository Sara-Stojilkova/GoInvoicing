import { useState } from "react";
import { useNavigate, useSearchParams, Link } from "react-router-dom";
import { Box, TextField, Button, Typography, Alert } from "@mui/material";
import { useAuth } from "../context/AuthContext";
import { ApiError } from "../api/error";

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
    <Box component="main" sx={{ maxWidth: 400, mx: "auto", mt: 8, p: 3 }}>
      <Typography variant="h5" component="h1" gutterBottom>
        Sign in
      </Typography>
      {registered && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Account created. Please sign in.
        </Alert>
      )}
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      <Box
        component="form"
        onSubmit={handleSubmit}
        noValidate
        sx={{ display: "flex", flexDirection: "column", gap: 2 }}
      >
        <TextField
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
          autoComplete="email"
        />
        <TextField
          label="Password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
          autoComplete="current-password"
        />
        <Button type="submit" variant="contained" disabled={loading}>
          {loading ? "Signing in…" : "Sign in"}
        </Button>
      </Box>
      <Typography sx={{ mt: 2 }}>
        Don't have an account? <Link to="/register">Register</Link>
      </Typography>
    </Box>
  );
}
