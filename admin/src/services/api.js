const API_BASE_URL = "/api";

class ApiClient {
  constructor() {
    this.baseUrl = API_BASE_URL;
  }

  getToken() {
    return localStorage.getItem("auth_token");
  }

  setToken(token) {
    if (token) {
      localStorage.setItem("auth_token", token);
    } else {
      localStorage.removeItem("auth_token");
    }
  }

  getHeaders() {
    const headers = { "Content-Type": "application/json" };
    const token = this.getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;
    return headers;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: { ...this.getHeaders(), ...options.headers },
    });

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Request failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // ==================== Auth ====================
  async login(data) {
    const result = await this.request("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
    if (result.success && result.data?.token) {
      this.setToken(result.data.token);
    }
    return result;
  }

  async getProfile() {
    return this.request("/auth/profile");
  }

  // ==================== Members ====================
  async getMembers(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/members?${query}`);
  }

  async getMember(id) {
    return this.request(`/members/${id}`);
  }

  async createMember(data) {
    return this.request("/members", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateMember(id, data) {
    return this.request(`/members/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteMember(id) {
    return this.request(`/members/${id}`, { method: "DELETE" });
  }

  async getFamilyMembers(memberId) {
    return this.request(`/members/${memberId}/family`);
  }

  async addFamilyMember(memberId, data) {
    return this.request(`/members/${memberId}/family`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateFamilyMember(memberId, familyId, data) {
    return this.request(`/members/${memberId}/family/${familyId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteFamilyMember(memberId, familyId) {
    return this.request(`/members/${memberId}/family/${familyId}`, {
      method: "DELETE",
    });
  }

  async resetMemberCredentials(memberId) {
    return this.request(`/members/${memberId}/reset-credentials`, {
      method: "POST",
    });
  }

  async sendMemberSms(memberId) {
    return this.request(`/members/${memberId}/send-sms`, { method: "POST" });
  }

  // ==================== Requests ====================
  async getRequests(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/requests${query ? `?${query}` : ""}`);
  }

  async getRequest(id) {
    return this.request(`/requests/${id}`);
  }

  async createRequest(data) {
    return this.request("/requests", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getMyRequests(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/requests/my${query ? `?${query}` : ""}`);
  }

  async approveRequest(id, data) {
    return this.request(`/requests/${id}/approve`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async rejectRequest(id, remarks) {
    return this.request(`/requests/${id}/reject`, {
      method: "POST",
      body: JSON.stringify({ remarks }),
    });
  }

  async getRequestStats(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/requests/statistics${query ? `?${query}` : ""}`);
  }

  // ==================== Payments ====================
  async getPayments(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/payments${query ? `?${query}` : ""}`);
  }

  async getPayment(id) {
    return this.request(`/payments/${id}`);
  }

  async createPayment(data) {
    return this.request("/payments", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getMyPayments(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/payments/my${query ? `?${query}` : ""}`);
  }

  async verifyPayment(id, data) {
    return this.request(`/payments/${id}/verify`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getPaymentStats(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/payments/statistics${query ? `?${query}` : ""}`);
  }
  // ==================== Transactions ====================
  async getTransactions(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/transactions${query ? `?${query}` : ""}`);
  }

  async getTransaction(id) {
    return this.request(`/transactions/${id}`);
  }

  async getTransactionSummary(fiscalYearId) {
    return this.request(`/transactions/summary?fiscal_year_id=${fiscalYearId}`);
  }

  async getMyTransactions(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/transactions/my${query ? `?${query}` : ""}`);
  }

  async getDashboardSummary() {
    return this.request("/transactions/dashboard-summary");
  }
  // ==================== Expenses ====================
  async getExpenses(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/expenses${query ? `?${query}` : ""}`);
  }

  async getExpense(id) {
    return this.request(`/expenses/${id}`);
  }

  async createExpense(data) {
    return this.request("/expenses", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateExpense(id, data) {
    return this.request(`/expenses/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteExpense(id) {
    return this.request(`/expenses/${id}`, { method: "DELETE" });
  }

  async uploadExpenseBillPhoto(id, formData) {
    const token = this.getToken();
    const response = await fetch(
      `${this.baseUrl}/expenses/${id}/upload-photo`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      },
    );

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // ==================== Expense Categories ====================
  async getExpenseCategories() {
    return this.request("/expenses/categories");
  }

  async getExpenseCategory(id) {
    return this.request(`/expenses/categories/${id}`);
  }

  async createExpenseCategory(data) {
    return this.request("/expenses/categories", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateExpenseCategory(id, data) {
    return this.request(`/expenses/categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteExpenseCategory(id) {
    return this.request(`/expenses/categories/${id}`, { method: "DELETE" });
  }

  // ==================== Fines ====================
  async getFines(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/fines${query ? `?${query}` : ""}`);
  }

  async getFine(id) {
    return this.request(`/fines/${id}`);
  }

  async createFine(data) {
    return this.request("/fines", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateFine(id, data) {
    return this.request(`/fines/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async updateFineStatus(id, data) {
    return this.request(`/fines/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }

  async deleteFine(id) {
    return this.request(`/fines/${id}`, { method: "DELETE" });
  }

  async getFineStats(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/fines/statistics${query ? `?${query}` : ""}`);
  }

  async uploadFinePhoto(id, formData) {
    const token = this.getToken();
    const response = await fetch(`${this.baseUrl}/fines/${id}/upload-photo`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: formData,
    });

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // ==================== Letters ====================
  async getLetters(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/letters${query ? `?${query}` : ""}`);
  }

  async getLetter(id) {
    return this.request(`/letters/${id}`);
  }

  async createLetter(data) {
    return this.request("/letters", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateLetter(id, data) {
    return this.request(`/letters/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteLetter(id) {
    return this.request(`/letters/${id}`, { method: "DELETE" });
  }

  async uploadLetterDocument(id, formData) {
    const token = this.getToken();
    const response = await fetch(
      `${this.baseUrl}/letters/${id}/upload-document`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      },
    );

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // ==================== Samiti Settings ====================
  async getSamitiSettings() {
    return this.request("/samiti/settings");
  }

  async updateSamitiSettings(data) {
    return this.request("/samiti/settings", {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async uploadSamitiLogo(formData) {
    const token = this.getToken();
    const response = await fetch(
      `${this.baseUrl}/samiti/settings/upload-logo`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      },
    );

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // ==================== Samiti Heads ====================
  async getSamitiHeads() {
    return this.request("/samiti/heads");
  }

  async getSamitiHead(id) {
    return this.request(`/samiti/heads/${id}`);
  }

  async createSamitiHead(data) {
    return this.request("/samiti/heads", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateSamitiHead(id, data) {
    return this.request(`/samiti/heads/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteSamitiHead(id) {
    return this.request(`/samiti/heads/${id}`, { method: "DELETE" });
  }

  async uploadHeadPhoto(id, formData) {
    const token = this.getToken();
    const response = await fetch(
      `${this.baseUrl}/samiti/heads/${id}/upload-photo`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      },
    );

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // ==================== Resources ====================
  // ==================== Resource Types ====================
  async getResourceTypes() {
    return this.request("/resources/types");
  }

  async getResourceType(id) {
    return this.request(`/resources/types/${id}`);
  }

  async createResourceType(data) {
    return this.request("/resources/types", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateResourceType(id, data) {
    return this.request(`/resources/types/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteResourceType(id) {
    return this.request(`/resources/types/${id}`, { method: "DELETE" });
  }

  // ==================== Resource Items ====================
  async getResourceItems(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/resources/items${query ? `?${query}` : ""}`);
  }

  async getResourceItem(id) {
    return this.request(`/resources/items/${id}`);
  }

  async createResourceItem(data) {
    return this.request("/resources/items", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateResourceItem(id, data) {
    return this.request(`/resources/items/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteResourceItem(id) {
    return this.request(`/resources/items/${id}`, { method: "DELETE" });
  }

  // ==================== Resource Rates ====================
  async getResourceRates(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/resources/rates${query ? `?${query}` : ""}`);
  }

  async getResourceRate(id) {
    return this.request(`/resources/rates/${id}`);
  }

  async createResourceRate(data) {
    return this.request("/resources/rates", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateResourceRate(id, data) {
    return this.request(`/resources/rates/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteResourceRate(id) {
    return this.request(`/resources/rates/${id}`, { method: "DELETE" });
  }

  // ==================== Stock ====================
  async getStock(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/resources/stock${query ? `?${query}` : ""}`);
  }

  async getStockItem(id) {
    return this.request(`/resources/stock/${id}`);
  }

  async createStock(data) {
    return this.request("/resources/stock", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateStock(id, data) {
    return this.request(`/resources/stock/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteStock(id) {
    return this.request(`/resources/stock/${id}`, { method: "DELETE" });
  }

  // ==================== Fiscal Years ====================
  async getFiscalYears(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/fiscal-years${query ? `?${query}` : ""}`);
  }

  async getFiscalYear(id) {
    return this.request(`/fiscal-years/${id}`);
  }

  async createFiscalYear(data) {
    return this.request("/fiscal-years", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateFiscalYear(id, data) {
    return this.request(`/fiscal-years/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteFiscalYear(id) {
    return this.request(`/fiscal-years/${id}`, { method: "DELETE" });
  }

  async activateFiscalYear(id) {
    return this.request(`/fiscal-years/${id}/set-active`, { method: "POST" });
  }

  // ==================== Fee Settings ====================
  async getFeeSettings(fiscalYearId) {
    return this.request(`/fiscal-years/${fiscalYearId}/fees`);
  }

  async createFeeSetting(fiscalYearId, data) {
    // Make sure the payload has the correct structure
    const payload = {
      fiscal_year_id: fiscalYearId,
      membership_fee: parseFloat(data.membership_fee),
    };
    return this.request("/fiscal-years/fee", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  }

  async updateFeeSetting(id, data) {
    const payload = {
      membership_fee: parseFloat(data.membership_fee),
    };
    return this.request(`/fiscal-years/fee/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    });
  }

  async deleteFeeSetting(id) {
    return this.request(`/fiscal-years/fee/${id}`, { method: "DELETE" });
  }
  // ==================== Reports ====================
  async getDashboard() {
    return this.request("/reports/dashboard");
  }

  async getDashboardCharts() {
    return this.request("/reports/dashboard/charts");
  }

  async getMemberReport() {
    return this.request("/reports/members");
  }

  async getResourceReport(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/reports/resources${query ? `?${query}` : ""}`);
  }

  async getFinancialReport(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/reports/financial${query ? `?${query}` : ""}`);
  }

  // Member Photo Uploads
  async uploadMemberPhoto(memberId, formData) {
    const token = this.getToken();
    const response = await fetch(
      `${this.baseUrl}/members/${memberId}/upload-photo`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      },
    );

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  async uploadAssistantPhoto(memberId, formData) {
    const token = this.getToken();
    const response = await fetch(
      `${this.baseUrl}/members/${memberId}/upload-assistant-photo`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      },
    );

    if (response.status === 401) {
      this.setToken(null);
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
      throw new Error("Unauthorized");
    }

    if (!response.ok) {
      const error = await response
        .json()
        .catch(() => ({ message: "Upload failed" }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }
}

export const api = new ApiClient();
