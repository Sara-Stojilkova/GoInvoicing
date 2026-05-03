import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  Box,
  TextField,
  Button,
  Typography,
  Alert,
  RadioGroup,
  FormControlLabel,
  Radio,
  FormLabel,
} from "@mui/material";
import { useAuth } from "../context/AuthContext";
import { ApiError } from "../api/error";

type AgencyMode = "create" | "join";

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
    <Box component="main" sx={{ maxWidth: 400, mx: "auto", mt: 8, p: 3 }}>
      <Typography variant="h5" component="h1" gutterBottom>
        Create account
      </Typography>
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
          label="Full name"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          required
          autoComplete="name"
        />
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
          autoComplete="new-password"
        />
        <Box>
          <FormLabel>Agency</FormLabel>
          <RadioGroup
            row
            value={agencyMode}
            onChange={(e) => setAgencyMode(e.target.value as AgencyMode)}
          >
            <FormControlLabel value="create" control={<Radio />} label="Create new" />
            <FormControlLabel value="join" control={<Radio />} label="Join existing" />
          </RadioGroup>
          {agencyMode === "create" ? (
            <TextField
              label="Agency name"
              value={agencyName}
              onChange={(e) => setAgencyName(e.target.value)}
              required
              fullWidth
            />
          ) : (
            <TextField
              label="Agency ID"
              value={agencyId}
              onChange={(e) => setAgencyId(e.target.value)}
              required
              fullWidth
              placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
            />
          )}
        </Box>
        <Button type="submit" variant="contained" disabled={loading}>
          {loading ? "Creating account…" : "Create account"}
        </Button>
      </Box>
      <Typography sx={{ mt: 2 }}>
        Already have an account? <Link to="/login">Sign in</Link>
      </Typography>
    </Box>
  );
}
