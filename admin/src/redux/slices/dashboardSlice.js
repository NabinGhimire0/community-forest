import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { api } from "../../services/api";

export const fetchDashboard = createAsyncThunk(
  "dashboard/fetchDashboard",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getDashboard();
      if (response.success) return response.data;
      return rejectWithValue(response.message || "Failed to load dashboard");
    } catch (error) {
      return rejectWithValue(error.message || "Failed to load dashboard");
    }
  },
);

export const fetchDashboardCharts = createAsyncThunk(
  "dashboard/fetchDashboardCharts",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getDashboardCharts();
      if (response.success) return response.data;
      return rejectWithValue(response.message || "Failed to load dashboard charts");
    } catch (error) {
      return rejectWithValue(error.message || "Failed to load dashboard charts");
    }
  },
);

const dashboardSlice = createSlice({
  name: "dashboard",
  initialState: {
    data: null,
    charts: null,
    dataStatus: "idle",
    chartStatus: "idle",
    error: null,
  },
  reducers: {
    clearDashboard(state) {
      state.data = null;
      state.charts = null;
      state.dataStatus = "idle";
      state.chartStatus = "idle";
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchDashboard.pending, (state) => {
        state.dataStatus = "loading";
        state.error = null;
      })
      .addCase(fetchDashboard.fulfilled, (state, action) => {
        state.dataStatus = "succeeded";
        state.data = action.payload;
      })
      .addCase(fetchDashboard.rejected, (state, action) => {
        state.dataStatus = "failed";
        state.error = action.payload;
      })
      .addCase(fetchDashboardCharts.pending, (state) => {
        state.chartStatus = "loading";
      })
      .addCase(fetchDashboardCharts.fulfilled, (state, action) => {
        state.chartStatus = "succeeded";
        state.charts = action.payload;
      })
      .addCase(fetchDashboardCharts.rejected, (state) => {
        state.chartStatus = "failed";
      });
  },
});

export const { clearDashboard } = dashboardSlice.actions;
export default dashboardSlice.reducer;
