const API_BASE_URL = (import.meta.env.VITE_API_URL || "/api").replace(/\/$/, "");

export class ApiError extends Error {
  constructor(message, { status = 0, code = "request_failed", payload = null } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.payload = payload;
  }
}

class ApiClient {
  constructor() {
    this.baseUrl = API_BASE_URL;
    this.csrfToken = "";
  }

  getCookie(name) {
    const prefix = `${encodeURIComponent(name)}=`;
    return document.cookie
      .split(";")
      .map((value) => value.trim())
      .find((value) => value.startsWith(prefix))
      ?.slice(prefix.length) || "";
  }

  isUnsafeMethod(method = "GET") {
    return !["GET", "HEAD", "OPTIONS"].includes(method.toUpperCase());
  }

  getHeaders(body, method = "GET") {
    const headers = {};
    if (!(body instanceof FormData) && body != null) {
      headers["Content-Type"] = "application/json";
    }
    if (this.isUnsafeMethod(method)) {
      const csrfToken = this.csrfToken || decodeURIComponent(this.getCookie("bansamiti_csrf"));
      if (csrfToken) headers["X-CSRF-Token"] = csrfToken;
    }
    return headers;
  }

  async fetchResponse(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const method = options.method || "GET";
    let response;
    try {
      response = await fetch(url, {
        ...options,
        method,
        credentials: "include",
        headers: {
          ...this.getHeaders(options.body, method),
          ...options.headers,
        },
      });
    } catch {
      throw new ApiError(
        "Cannot connect to the server. Make sure the backend is running and the API URL is correct.",
        { code: "network_error" },
      );
    }
    const refreshedCSRF = response.headers.get("X-CSRF-Token");
    if (refreshedCSRF) this.csrfToken = refreshedCSRF;
    return response;
  }

  async request(endpoint, options = {}) {
    const response = await this.fetchResponse(endpoint, options);
    const payload = await response.json().catch(() => null);

    if (response.status === 401 && endpoint !== "/auth/login") {
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
    }
    if (!response.ok) {
      throw new ApiError(
        payload?.message || payload?.error || `Request failed with status ${response.status}`,
        {
          status: response.status,
          code: payload?.code || "request_failed",
          payload,
        },
      );
    }
    return payload;
  }

  async uploadForm(endpoint, formData) {
    return this.request(endpoint, { method: "POST", body: formData });
  }

  async requestBlob(endpoint, options = {}) {
    const response = await this.fetchResponse(endpoint, options);
    if (!response.ok) {
      const payload = await response.json().catch(() => null);
      throw new ApiError(
        payload?.message || payload?.error || `Download failed with status ${response.status}`,
        { status: response.status, code: payload?.code || "download_failed", payload },
      );
    }
    const disposition = response.headers.get("Content-Disposition") || "";
    const utf8Match = disposition.match(/filename\*=UTF-8''([^;]+)/i);
    const normalMatch = disposition.match(/filename="?([^";]+)"?/i);
    const filename = decodeURIComponent(utf8Match?.[1] || normalMatch?.[1] || "download.bin");
    return { blob: await response.blob(), filename };
  }

  saveDownload({ blob, filename }) {
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }

  // ==================== Connection ====================
  async health() {
    return this.request("/health");
  }

  // ==================== Auth ====================
  async login(data) {
    return this.request("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async logout() {
    try {
      return await this.request("/auth/logout", { method: "POST" });
    } finally {
      this.csrfToken = "";
    }
  }

  async getProfile() {
    return this.request("/auth/profile");
  }

  async changePassword(data) {
    return this.request("/auth/change-password", { method: "PUT", body: JSON.stringify(data) });
  }

  async getSessions() {
    return this.request("/auth/sessions");
  }

  async revokeAllSessions() {
    return this.request("/auth/sessions/revoke-all", { method: "POST" });
  }

  async beginMFA(currentPassword) {
    return this.request("/auth/mfa/setup", { method: "POST", body: JSON.stringify({ current_password: currentPassword }) });
  }

  async enableMFA(code) {
    return this.request("/auth/mfa/enable", { method: "POST", body: JSON.stringify({ code }) });
  }

  async disableMFA(currentPassword, code) {
    return this.request("/auth/mfa/disable", { method: "POST", body: JSON.stringify({ current_password: currentPassword, code }) });
  }

  // ==================== Members ====================
  async getMembers(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/members${query ? `?${query}` : ""}`);
  }

  async getMember(id) {
    return this.request(`/members/${id}`);
  }

  async getMyMemberProfile() {
    return this.request("/members/profile");
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

  async resetMemberCredentials(memberId, security) {
    return this.request(`/members/${memberId}/reset-credentials`, {
      method: "POST",
      body: JSON.stringify(security),
    });
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

  async createCashPayment(data) {
    return this.request("/payments/cash", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async initiateEsewaPayment(requestId) {
    return this.request("/payments/esewa/initiate", {
      method: "POST",
      body: JSON.stringify({ request_id: Number(requestId) }),
    });
  }

  async initiateEsewaLedgerPayment(ledgerTransactionId) {
    return this.request("/payments/esewa/initiate", {
      method: "POST",
      body: JSON.stringify({
        ledger_transaction_id: Number(ledgerTransactionId),
      }),
    });
  }

  submitEsewaForm(actionUrl, fields) {
    const form = document.createElement("form");
    form.method = "POST";
    form.action = actionUrl;
    Object.entries(fields || {}).forEach(([name, value]) => {
      const input = document.createElement("input");
      input.type = "hidden";
      input.name = name;
      input.value = String(value);
      form.appendChild(input);
    });
    document.body.appendChild(form);
    form.submit();
  }

  async checkEsewaPaymentStatus(paymentId) {
    return this.request(`/payments/esewa/${paymentId}/check-status`, {
      method: "POST",
    });
  }

  async getMyPayments(params = {}) {
    const query = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v != null),
    ).toString();
    return this.request(`/payments/my${query ? `?${query}` : ""}`);
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
    return this.uploadForm(`/expenses/${id}/upload-photo`, formData);
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
    return this.uploadForm(`/fines/${id}/upload-photo`, formData);
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
    return this.uploadForm(`/letters/${id}/upload-document`, formData);
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
    return this.uploadForm("/samiti/settings/upload-logo", formData);
  }

  // ==================== Samiti Heads ====================
  async getSamitiHeads() {
    return this.request("/samiti/heads");
  }

  async getSamitiHead(id) {
    return this.request(`/samiti/heads/${id}`);
  }

  async getManagedSamitiHeads() {
    return this.request("/samiti/manage/heads");
  }

  async getManagedSamitiHead(id) {
    return this.request(`/samiti/manage/heads/${id}`);
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
    return this.uploadForm(`/samiti/heads/${id}/upload-photo`, formData);
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
  // ==================== Historical ledger ====================
  async createHistoricalTransaction(memberId, data) {
    return this.request(`/members/${memberId}/historical-transaction`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async verifyHistoricalTransaction(transactionId) {
    return this.request(`/members/historical-transactions/${transactionId}/verify`, {
      method: "POST",
    });
  }

  async reverseHistoricalTransaction(transactionId, reason) {
    return this.request(`/members/historical-transactions/${transactionId}/reverse`, {
      method: "POST",
      body: JSON.stringify({ reason }),
    });
  }

  async uploadTransactionDocument(transactionId, file) {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("folder", "documents");
    formData.append("entity", "transaction");
    formData.append("entity_id", String(transactionId));
    return this.request("/uploads", { method: "POST", body: formData });
  }

  async downloadAuthenticated(endpoint, fallbackFilename = "download") {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      headers: this.getHeaders(null),
    });
    if (!response.ok) {
      const payload = await response.json().catch(() => null);
      throw new Error(payload?.message || "Download failed");
    }
    const blob = await response.blob();
    const disposition = response.headers.get("content-disposition") || "";
    const match = disposition.match(/filename="?([^";]+)"?/i);
    const filename = match?.[1] || fallbackFilename;
    const href = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = href;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(href);
  }

  async downloadPaymentReceipt(paymentId) {
    return this.downloadAuthenticated(`/receipts/payment/${paymentId}`, `payment-${paymentId}.pdf`);
  }

  async downloadUploadedFile(file) {
    if (!file?.file_url) throw new Error("Document is unavailable");
    const endpoint = file.file_url.startsWith(this.baseUrl)
      ? file.file_url.slice(this.baseUrl.length)
      : file.file_url.replace(/^\/api/, "");
    return this.downloadAuthenticated(endpoint, file.original_name || "document");
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
    return this.uploadForm(`/members/${memberId}/upload-photo`, formData);
  }

  async uploadAssistantPhoto(memberId, formData) {
    return this.uploadForm(`/members/${memberId}/upload-assistant-photo`, formData);
  }
  async getMemberFinancialSummary(memberId) {
    return this.request(`/members/${memberId}/financial-summary`);
  }

  // Add these to ApiClient class
  async getMemberFeeDetails(memberId) {
    return this.request(`/members/${memberId}/fee-details`);
  }

  async getMemberSalesDetails(memberId) {
    return this.request(`/members/${memberId}/sales-details`);
  }

  // ==================== Production data operations (Admin) ====================
  async getExportDatasets() {
    return this.request("/admin/system/exports/datasets");
  }

  async exportDataset(data) {
    return this.requestBlob("/admin/system/exports/csv", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async exportAllData(data) {
    return this.requestBlob("/admin/system/exports/all", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async createDatabaseBackup(data) {
    return this.requestBlob("/admin/system/backups/database", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async createFullBackup(data) {
    return this.requestBlob("/admin/system/backups/full", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }
}

export const api = new ApiClient();
