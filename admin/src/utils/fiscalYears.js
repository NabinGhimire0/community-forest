export function getActiveFiscalYear(fiscalYears = []) {
  return fiscalYears.find((fiscalYear) => fiscalYear?.is_active) || null;
}

export function getActiveFiscalYearId(fiscalYears = []) {
  const activeFiscalYear = getActiveFiscalYear(fiscalYears);
  return activeFiscalYear ? String(activeFiscalYear.id) : "";
}

export function buildFiscalYearOptions(fiscalYears = [], { includeAll = false } = {}) {
  const options = fiscalYears.map((fiscalYear) => ({
    value: String(fiscalYear.id),
    label: `${fiscalYear.name}${fiscalYear.is_active ? " (Active)" : ""}`,
  }));

  return includeAll
    ? [{ value: "", label: "All Fiscal Years" }, ...options]
    : options;
}
