import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api } from "../../services/api";

export const DEFAULT_APP_SETTINGS = {
  name: "Community Forestry Management System",
  registration_no: null,
  address: "",
  ward_no: null,
  municipality: "",
  district: "",
  province: "",
  contact_phone: null,
  contact_email: null,
  description: null,
  logo: null,
  established_date: null,
};


const readCachedSettings = () => {
  try {
    const cached = localStorage.getItem("app_settings");
    return cached
      ? { ...DEFAULT_APP_SETTINGS, ...JSON.parse(cached) }
      : DEFAULT_APP_SETTINGS;
  } catch {
    return DEFAULT_APP_SETTINGS;
  }
};

const cacheSettings = (settings) => {
  try {
    localStorage.setItem("app_settings", JSON.stringify(settings));
  } catch {
    // The UI can still use the in-memory settings when storage is unavailable.
  }
};

export const fetchAppSettings = createAsyncThunk(
  "appSettings/fetchSettings",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getSamitiSettings();
      if (response.success && response.data) return response.data;
      return rejectWithValue(response.message || "Failed to load organization settings");
    } catch (error) {
      return rejectWithValue(error.message || "Failed to load organization settings");
    }
  },
);

export const fetchPublicCommittee = createAsyncThunk(
  "appSettings/fetchCommittee",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getSamitiHeads();
      if (response.success) return response.data || [];
      return rejectWithValue(response.message || "Failed to load committee members");
    } catch (error) {
      return rejectWithValue(error.message || "Failed to load committee members");
    }
  },
);

const appSettingsSlice = createSlice({
  name: "appSettings",
  initialState: {
    settings: readCachedSettings(),
    committee: [],
    status: "idle",
    committeeStatus: "idle",
    error: null,
  },
  reducers: {
    setAppSettings(state, action) {
      state.settings = { ...DEFAULT_APP_SETTINGS, ...action.payload };
      cacheSettings(state.settings);
      state.status = "succeeded";
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchAppSettings.pending, (state) => {
        state.status = "loading";
        state.error = null;
      })
      .addCase(fetchAppSettings.fulfilled, (state, action) => {
        state.status = "succeeded";
        state.settings = { ...DEFAULT_APP_SETTINGS, ...action.payload };
        cacheSettings(state.settings);
      })
      .addCase(fetchAppSettings.rejected, (state, action) => {
        state.status = "failed";
        state.error = action.payload;
      })
      .addCase(fetchPublicCommittee.pending, (state) => {
        state.committeeStatus = "loading";
      })
      .addCase(fetchPublicCommittee.fulfilled, (state, action) => {
        state.committeeStatus = "succeeded";
        state.committee = action.payload;
      })
      .addCase(fetchPublicCommittee.rejected, (state) => {
        state.committeeStatus = "failed";
      });
  },
});

export const { setAppSettings } = appSettingsSlice.actions;
export default appSettingsSlice.reducer;
