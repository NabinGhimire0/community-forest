import { createSlice, createAsyncThunk } from "@reduxjs/toolkit";
import { api } from "../../services/api";

export const fetchDashboard = createAsyncThunk(
  "dashboard/fetchDashboard",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getDashboard();
      if (response.success) return response.data;
      return rejectWithValue(response.message);
    } catch (error) {
      return rejectWithValue(error.message);
    }
  },
);

export const fetchDashboardCharts = createAsyncThunk(
  "dashboard/fetchDashboardCharts",
  async (_, { rejectWithValue }) => {
    try {
      const response = await api.getDashboardCharts();
      if (response.success) return response.data;
      return rejectWithValue(response.message);
    } catch (error) {
      return rejectWithValue(error.message);
    }
  },
);

const dashboardSlice = createSlice({
  name: "dashboard",
  initialState: {
    data: null,
    charts: null,
    isLoading: false,
    error: null,
  },
  reducers: {},
  extraReducers: (builder) => {
    builder
      // Dashboard data
      .addCase(fetchDashboard.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchDashboard.fulfilled, (state, action) => {
        state.isLoading = false;
        state.data = action.payload;
      })
      .addCase(fetchDashboard.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload;
      })
      // Dashboard charts
      .addCase(fetchDashboardCharts.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(fetchDashboardCharts.fulfilled, (state, action) => {
        state.isLoading = false;
        state.charts = action.payload;
      })
      .addCase(fetchDashboardCharts.rejected, (state) => {
        state.isLoading = false;
      });
  },
});

export default dashboardSlice.reducer;
