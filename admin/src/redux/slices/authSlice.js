import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api } from "../../services/api";

const rejection = (error, fallback) => ({
  message: error?.message || fallback,
  code: error?.code || "request_failed",
});

export const login = createAsyncThunk(
  "auth/login",
  async ({ phone, password, otp = "" }, { rejectWithValue }) => {
    try {
      const response = await api.login({ phone, password, otp });
      if (response.success && response.data?.user) return response.data;
      return rejectWithValue({ message: response.message || "Login failed", code: "login_failed" });
    } catch (error) {
      return rejectWithValue(rejection(error, "Login failed"));
    }
  },
);

export const loadProfile = createAsyncThunk(
  "auth/loadProfile",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getProfile();
      if (response.success && response.data) return response.data;
      return rejectWithValue({ message: "Failed to load profile", code: "profile_failed" });
    } catch (error) {
      return rejectWithValue(rejection(error, "Failed to load profile"));
    }
  },
);

export const logout = createAsyncThunk("auth/logout", async () => {
  try {
    await api.logout();
  } catch {
    // Clear the local authenticated view even when the network is unavailable.
  }
});

const initialState = {
  user: null,
  isAuthenticated: false,
  isLoading: false,
  isProfileLoading: true,
  hasProfileLoaded: false,
  error: null,
  errorCode: null,
};

const clearSessionState = (state) => {
  state.user = null;
  state.isAuthenticated = false;
  state.isLoading = false;
  state.isProfileLoading = false;
  state.hasProfileLoaded = true;
  state.error = null;
  state.errorCode = null;
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    sessionExpired(state) {
      clearSessionState(state);
    },
    clearError(state) {
      state.error = null;
      state.errorCode = null;
    },
    updateAuthenticatedUser(state, action) {
      state.user = { ...state.user, ...action.payload };
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => {
        state.isLoading = true;
        state.error = null;
        state.errorCode = null;
      })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload.user;
        state.isAuthenticated = true;
        state.isProfileLoading = false;
        state.hasProfileLoaded = true;
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload?.message || "Login failed";
        state.errorCode = action.payload?.code || "login_failed";
      })
      .addCase(loadProfile.pending, (state) => {
        state.isProfileLoading = true;
      })
      .addCase(loadProfile.fulfilled, (state, action) => {
        state.user = action.payload;
        state.isAuthenticated = true;
        state.isProfileLoading = false;
        state.hasProfileLoaded = true;
      })
      .addCase(loadProfile.rejected, (state) => {
        clearSessionState(state);
      })
      .addCase(logout.fulfilled, (state) => {
        clearSessionState(state);
      });
  },
});

export const { sessionExpired, clearError, updateAuthenticatedUser } = authSlice.actions;
export default authSlice.reducer;
