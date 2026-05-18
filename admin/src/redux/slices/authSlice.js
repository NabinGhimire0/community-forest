import { createSlice, createAsyncThunk } from "@reduxjs/toolkit";
import { api } from "../../services/api";

export const login = createAsyncThunk(
  "auth/login",
  async ({ phone, password }, { rejectWithValue }) => {
    try {
      const response = await api.login({ phone, password });
      if (response.success && response.data) {
        localStorage.setItem("auth_user", JSON.stringify(response.data.user));
        return response.data;
      }
      return rejectWithValue(response.message || "Login failed");
    } catch (error) {
      return rejectWithValue(error.message || "Login failed");
    }
  },
);

export const loadProfile = createAsyncThunk(
  "auth/loadProfile",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getProfile();
      if (response.success && response.data) {
        localStorage.setItem("auth_user", JSON.stringify(response.data));
        return response.data;
      }
      return rejectWithValue("Failed to load profile");
    } catch (error) {
      return rejectWithValue(error.message);
    }
  },
);

const getInitialAuthState = () => {
  try {
    const token = localStorage.getItem("auth_token");
    const user = localStorage.getItem("auth_user");
    return {
      user: user ? JSON.parse(user) : null,
      token: token || null,
      isAuthenticated: !!token,
      isLoading: false,
      error: null,
    };
  } catch {
    return {
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    };
  }
};

const authSlice = createSlice({
  name: "auth",
  initialState: getInitialAuthState(),
  reducers: {
    logout(state) {
      state.user = null;
      state.token = null;
      state.isAuthenticated = false;
      state.error = null;
      localStorage.removeItem("auth_token");
      localStorage.removeItem("auth_user");
      api.setToken(null);
    },
    clearError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload.user;
        state.token = action.payload.token;
        state.isAuthenticated = true;
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload || "Login failed";
      })
      .addCase(loadProfile.fulfilled, (state, action) => {
        state.user = action.payload;
        state.isAuthenticated = true;
      })
      .addCase(loadProfile.rejected, (state) => {
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
        localStorage.removeItem("auth_token");
        localStorage.removeItem("auth_user");
      });
  },
});

export const { logout, clearError } = authSlice.actions;
export default authSlice.reducer;
