// ─── Members ────────────────────────────────────────────
export const members = [
  { id: 1, name: 'Ram Bahadur Shrestha', phone: '9841234567', ward: 1, address: 'Bhaktapur-1, Ward No. 1', status: 'active', joinedDate: '2020-01-15', shareAmount: 5000 },
  { id: 2, name: 'Sita Kumari Thapa', phone: '9851234568', ward: 1, address: 'Bhaktapur-1, Ward No. 1', status: 'active', joinedDate: '2020-03-10', shareAmount: 5000 },
  { id: 3, name: 'Hari Prasad Adhikari', phone: '9861234569', ward: 2, address: 'Bhaktapur-2, Ward No. 2', status: 'active', joinedDate: '2020-05-22', shareAmount: 5000 },
  { id: 4, name: 'Gita Devi Pokharel', phone: '9871234570', ward: 2, address: 'Bhaktapur-2, Ward No. 2', status: 'inactive', joinedDate: '2019-11-08', shareAmount: 5000 },
  { id: 5, name: 'Krishna Lal Maharjan', phone: '9801234571', ward: 3, address: 'Bhaktapur-3, Ward No. 3', status: 'active', joinedDate: '2021-02-14', shareAmount: 5000 },
  { id: 6, name: 'Laxmi Maya Tamang', phone: '9811234572', ward: 3, address: 'Bhaktapur-3, Ward No. 3', status: 'active', joinedDate: '2021-06-30', shareAmount: 5000 },
  { id: 7, name: 'Bhim Bahadur Gurung', phone: '9821234573', ward: 4, address: 'Bhaktapur-4, Ward No. 4', status: 'active', joinedDate: '2019-08-19', shareAmount: 5000 },
  { id: 8, name: 'Saraswoti Rai', phone: '9831234574', ward: 4, address: 'Bhaktapur-4, Ward No. 4', status: 'active', joinedDate: '2022-01-05', shareAmount: 5000 },
  { id: 9, name: 'Dipak Kumar Singh', phone: '9841234575', ward: 5, address: 'Bhaktapur-5, Ward No. 5', status: 'pending', joinedDate: '2024-01-20', shareAmount: 5000 },
  { id: 10, name: 'Nirmala Devi Sharma', phone: '9851234576', ward: 5, address: 'Bhaktapur-5, Ward No. 5', status: 'active', joinedDate: '2020-09-12', shareAmount: 5000 },
  { id: 11, name: 'Mohan Prasad Koirala', phone: '9861234577', ward: 1, address: 'Bhaktapur-1, Ward No. 1', status: 'active', joinedDate: '2019-04-28', shareAmount: 5000 },
  { id: 12, name: 'Pushpa Kumari Bhandari', phone: '9871234578', ward: 2, address: 'Bhaktapur-2, Ward No. 2', status: 'active', joinedDate: '2021-11-15', shareAmount: 5000 },
];

// ─── Requests ───────────────────────────────────────────
export const requests = [
  { id: 1, memberId: 1, memberName: 'Ram Bahadur Shrestha', resourceItemName: 'Sal Timber', resourceTypeName: 'Timber', quantity: 50, unit: 'cubic feet', purpose: 'House construction', status: 'approved', requestDate: '2024-01-10', approvedDate: '2024-01-12', approvedBy: 'Admin' },
  { id: 2, memberId: 2, memberName: 'Sita Kumari Thapa', resourceItemName: 'Sal Timber', resourceTypeName: 'Timber', quantity: 30, unit: 'cubic feet', purpose: 'Furniture', status: 'pending', requestDate: '2024-01-15', approvedDate: null, approvedBy: null },
  { id: 3, memberId: 3, memberName: 'Hari Prasad Adhikari', resourceItemName: 'Firewood', resourceTypeName: 'Firewood', quantity: 200, unit: 'kg', purpose: 'Domestic use', status: 'fulfilled', requestDate: '2024-01-08', approvedDate: '2024-01-09', approvedBy: 'Admin' },
  { id: 4, memberId: 5, memberName: 'Krishna Lal Maharjan', resourceItemName: 'Bamboo', resourceTypeName: 'NTFP', quantity: 100, unit: 'pieces', purpose: 'Scaffolding', status: 'rejected', requestDate: '2024-01-12', approvedDate: null, approvedBy: null },
  { id: 5, memberId: 6, memberName: 'Laxmi Maya Tamang', resourceItemName: 'Sal Timber', resourceTypeName: 'Timber', quantity: 80, unit: 'cubic feet', purpose: 'House repair', status: 'approved', requestDate: '2024-01-18', approvedDate: '2024-01-20', approvedBy: 'Staff' },
  { id: 6, memberId: 7, memberName: 'Bhim Bahadur Gurung', resourceItemName: 'Firewood', resourceTypeName: 'Firewood', quantity: 500, unit: 'kg', purpose: 'Commercial', status: 'pending', requestDate: '2024-01-22', approvedDate: null, approvedBy: null },
  { id: 7, memberId: 8, memberName: 'Saraswoti Rai', resourceItemName: 'Khair', resourceTypeName: 'Timber', quantity: 20, unit: 'cubic feet', purpose: 'Agricultural tools', status: 'approved', requestDate: '2024-01-05', approvedDate: '2024-01-06', approvedBy: 'Admin' },
  { id: 8, memberId: 10, memberName: 'Nirmala Devi Sharma', resourceItemName: 'Sal Leaf', resourceTypeName: 'NTFP', quantity: 5000, unit: 'pieces', purpose: 'Plate making', status: 'fulfilled', requestDate: '2024-01-20', approvedDate: '2024-01-21', approvedBy: 'Staff' },
];

// ─── Resource Types ─────────────────────────────────────
export const resourceTypes = [
  { id: 1, name: 'Timber', description: 'All types of timber products', active: true },
  { id: 2, name: 'Firewood', description: 'Firewood and fuel products', active: true },
  { id: 3, name: 'NTFP', description: 'Non-Timber Forest Products', active: true },
  { id: 4, name: 'Grass & Fodder', description: 'Grass and animal fodder', active: true },
  { id: 5, name: 'Stone & Sand', description: 'Construction materials from forest', active: false },
];

// ─── Resource Items ─────────────────────────────────────
export const resourceItems = [
  { id: 1, typeId: 1, typeName: 'Timber', name: 'Sal Timber', unit: 'cubic feet', description: 'Sal wood - primary construction timber' },
  { id: 2, typeId: 1, typeName: 'Timber', name: 'Khair', unit: 'cubic feet', description: 'Khair wood - tools and agriculture' },
  { id: 3, typeId: 1, typeName: 'Timber', name: 'Sissoo', unit: 'cubic feet', description: 'Sissoo wood - furniture grade' },
  { id: 4, typeId: 2, typeName: 'Firewood', name: 'Firewood', unit: 'kg', description: 'Mixed firewood bundle' },
  { id: 5, typeId: 3, typeName: 'NTFP', name: 'Bamboo', unit: 'pieces', description: 'Bamboo poles' },
  { id: 6, typeId: 3, typeName: 'NTFP', name: 'Sal Leaf', unit: 'pieces', description: 'Dry sal leaves for plates' },
  { id: 7, typeId: 3, typeName: 'NTFP', name: 'Mushroom (Seeds)', unit: 'packets', description: 'Oyster mushroom seed packets' },
  { id: 8, typeId: 4, typeName: 'Grass & Fodder', name: 'Napier Grass', unit: 'kg', description: 'Animal feed grass' },
];

// ─── Resource Rates ─────────────────────────────────────
export const resourceRates = [
  { id: 1, itemId: 1, itemName: 'Sal Timber', rate: 800, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 2, itemId: 2, itemName: 'Khair', rate: 600, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 3, itemId: 3, itemName: 'Sissoo', rate: 1200, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 4, itemId: 4, itemName: 'Firewood', rate: 15, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 5, itemId: 5, itemName: 'Bamboo', rate: 50, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 6, itemId: 6, itemName: 'Sal Leaf', rate: 2, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 7, itemId: 7, itemName: 'Mushroom (Seeds)', rate: 100, effectiveDate: '2081-01-01', fiscalYearId: 1 },
  { id: 8, itemId: 8, itemName: 'Napier Grass', rate: 5, effectiveDate: '2081-01-01', fiscalYearId: 1 },
];

// ─── Resource Stock ─────────────────────────────────────
export const resourceStock = [
  { id: 1, itemId: 1, itemName: 'Sal Timber', quantity: 2500, unit: 'cubic feet', distributed: 480 },
  { id: 2, itemId: 2, itemName: 'Khair', quantity: 800, unit: 'cubic feet', distributed: 120 },
  { id: 3, itemId: 3, itemName: 'Sissoo', quantity: 400, unit: 'cubic feet', distributed: 60 },
  { id: 4, itemId: 4, itemName: 'Firewood', quantity: 15000, unit: 'kg', distributed: 3500 },
  { id: 5, itemId: 5, itemName: 'Bamboo', quantity: 2000, unit: 'pieces', distributed: 350 },
  { id: 6, itemId: 6, itemName: 'Sal Leaf', quantity: 50000, unit: 'pieces', distributed: 12000 },
  { id: 7, itemId: 7, itemName: 'Mushroom (Seeds)', quantity: 200, unit: 'packets', distributed: 45 },
  { id: 8, itemId: 8, itemName: 'Napier Grass', quantity: 8000, unit: 'kg', distributed: 2100 },
];

// ─── Payments ───────────────────────────────────────────
export const payments = [
  { id: 1, requestId: 1, memberName: 'Ram Bahadur Shrestha', itemName: 'Sal Timber', quantity: 50, unit: 'cubic feet', rate: 800, totalAmount: 40000, paymentMethod: 'cash', status: 'paid', paymentDate: '2024-01-14', receiptNo: 'PAY-2024-001' },
  { id: 2, requestId: 3, memberName: 'Hari Prasad Adhikari', itemName: 'Firewood', quantity: 200, unit: 'kg', rate: 15, totalAmount: 3000, paymentMethod: 'bank', status: 'paid', paymentDate: '2024-01-10', receiptNo: 'PAY-2024-002' },
  { id: 3, requestId: 5, memberName: 'Laxmi Maya Tamang', itemName: 'Sal Timber', quantity: 80, unit: 'cubic feet', rate: 800, totalAmount: 64000, paymentMethod: 'cash', status: 'partially_paid', paymentDate: '2024-01-22', receiptNo: 'PAY-2024-003', paidAmount: 32000 },
  { id: 4, requestId: 7, memberName: 'Saraswoti Rai', itemName: 'Khair', quantity: 20, unit: 'cubic feet', rate: 600, totalAmount: 12000, paymentMethod: 'bank', status: 'paid', paymentDate: '2024-01-08', receiptNo: 'PAY-2024-004' },
  { id: 5, requestId: 8, memberName: 'Nirmala Devi Sharma', itemName: 'Sal Leaf', quantity: 5000, unit: 'pieces', rate: 2, totalAmount: 10000, paymentMethod: 'cash', status: 'paid', paymentDate: '2024-01-23', receiptNo: 'PAY-2024-005' },
  { id: 6, requestId: 2, memberName: 'Sita Kumari Thapa', itemName: 'Sal Timber', quantity: 30, unit: 'cubic feet', rate: 800, totalAmount: 24000, paymentMethod: 'cash', status: 'unpaid', paymentDate: null, receiptNo: null },
];

// ─── Transactions ───────────────────────────────────────
export const transactions = [
  { id: 1, date: '2024-01-14', description: 'Timber sale - Ram Bahadur Shrestha', type: 'income', category: 'Resource Sale', amount: 40000 },
  { id: 2, date: '2024-01-10', description: 'Firewood sale - Hari Prasad Adhikari', type: 'income', category: 'Resource Sale', amount: 3000 },
  { id: 3, date: '2024-01-15', description: 'Forest patrol salary - January', type: 'expense', category: 'Salary', amount: 15000 },
  { id: 4, date: '2024-01-16', description: 'Seedling purchase', type: 'expense', category: 'Plantation', amount: 8000 },
  { id: 5, date: '2024-01-08', description: 'Khair sale - Saraswoti Rai', type: 'income', category: 'Resource Sale', amount: 12000 },
  { id: 6, date: '2024-01-20', description: 'Office rent - January', type: 'expense', category: 'Administrative', amount: 5000 },
  { id: 7, date: '2024-01-22', description: 'Partial payment - Laxmi Maya Tamang', type: 'income', category: 'Resource Sale', amount: 32000 },
  { id: 8, date: '2024-01-23', description: 'Sal leaf sale - Nirmala Devi Sharma', type: 'income', category: 'Resource Sale', amount: 10000 },
  { id: 9, date: '2024-01-25', description: 'Equipment maintenance', type: 'expense', category: 'Maintenance', amount: 3500 },
  { id: 10, date: '2024-01-28', description: 'Community development fund', type: 'expense', category: 'Community', amount: 20000 },
];

// ─── Expenses ───────────────────────────────────────────
export const expenses = [
  { id: 1, date: '2024-01-15', description: 'Forest patrol salary - January', category: 'Salary', amount: 15000, approvedBy: 'Admin' },
  { id: 2, date: '2024-01-16', description: 'Seedling purchase for plantation', category: 'Plantation', amount: 8000, approvedBy: 'Admin' },
  { id: 3, date: '2024-01-20', description: 'Office rent - January', category: 'Administrative', amount: 5000, approvedBy: 'Staff' },
  { id: 4, date: '2024-01-25', description: 'Equipment repair and maintenance', category: 'Maintenance', amount: 3500, approvedBy: 'Admin' },
  { id: 5, date: '2024-01-28', description: 'Community development contribution', category: 'Community', amount: 20000, approvedBy: 'Admin' },
  { id: 6, date: '2024-02-01', description: 'Meeting refreshment cost', category: 'Meeting', amount: 2500, approvedBy: 'Staff' },
];

// ─── Fines ──────────────────────────────────────────────
export const fines = [
  { id: 1, memberId: 4, memberName: 'Gita Devi Pokharel', amount: 5000, reason: 'Unauthorized tree cutting', status: 'paid', fineDate: '2023-12-15', paidDate: '2023-12-20' },
  { id: 2, memberId: 9, memberName: 'Dipak Kumar Singh', amount: 3000, reason: 'Over-extraction of resources', status: 'unpaid', fineDate: '2024-01-10', paidDate: null },
  { id: 3, memberId: 3, memberName: 'Hari Prasad Adhikari', amount: 2000, reason: 'Late payment penalty', status: 'paid', fineDate: '2024-01-05', paidDate: '2024-01-08' },
  { id: 4, memberId: 6, memberName: 'Laxmi Maya Tamang', amount: 1500, reason: 'Cattle grazing violation', status: 'unpaid', fineDate: '2024-01-18', paidDate: null },
];

// ─── Letters ────────────────────────────────────────────
export const letters = [
  { id: 1, referenceNo: 'BS-LET-2024-001', to: 'District Forest Office', subject: 'Annual Progress Report Submission', type: 'Official', issueDate: '2024-01-15', status: 'sent' },
  { id: 2, referenceNo: 'BS-LET-2024-002', to: 'Ward Office - Ward 3', subject: 'Forest Boundary Demarcation Request', type: 'Request', issueDate: '2024-01-18', status: 'sent' },
  { id: 3, referenceNo: 'BS-LET-2024-003', to: 'All Members', subject: 'Annual General Meeting Notice', type: 'Notice', issueDate: '2024-01-20', status: 'draft' },
  { id: 4, referenceNo: 'BS-LET-2024-004', to: 'DFO Bhaktapur', subject: 'Firewood Collection Permission', type: 'Permission', issueDate: '2024-01-22', status: 'approved' },
];

// ─── Samiti Heads (Committee) ──────────────────────────
export const samitiHeads = [
  { id: 1, memberId: 11, memberName: 'Mohan Prasad Koirala', position: 'Chairman', phone: '9861234577', termStart: '2081-01-01', termEnd: '2083-12-30', status: 'active' },
  { id: 2, memberId: 1, memberName: 'Ram Bahadur Shrestha', position: 'Vice Chairman', phone: '9841234567', termStart: '2081-01-01', termEnd: '2083-12-30', status: 'active' },
  { id: 3, memberId: 3, memberName: 'Hari Prasad Adhikari', position: 'Secretary', phone: '9861234569', termStart: '2081-01-01', termEnd: '2083-12-30', status: 'active' },
  { id: 4, memberId: 7, memberName: 'Bhim Bahadur Gurung', position: 'Treasurer', phone: '9821234573', termStart: '2081-01-01', termEnd: '2083-12-30', status: 'active' },
  { id: 5, memberId: 2, memberName: 'Sita Kumari Thapa', position: 'Member', phone: '9851234568', termStart: '2081-01-01', termEnd: '2083-12-30', status: 'active' },
  { id: 6, memberId: 8, memberName: 'Saraswoti Rai', position: 'Member', phone: '9831234574', termStart: '2081-01-01', termEnd: '2083-12-30', status: 'active' },
];

// ─── KPI Stats ──────────────────────────────────────────
export const kpiStats = [
  { title: 'Total Members', value: '156', change: '+12', trend: 'up', description: 'from last quarter', icon: 'Users' },
  { title: 'Total Revenue', value: 'NPR 1.2M', change: '+18.5%', trend: 'up', description: 'this fiscal year', icon: 'DollarSign' },
  { title: 'Pending Requests', value: '23', change: '-5', trend: 'down', description: 'awaiting approval', icon: 'FileText' },
  { title: 'Active Resources', value: '8', change: '+2', trend: 'up', description: 'resource types in stock', icon: 'Package' },
];

// ─── Revenue Data (monthly) ────────────────────────────
export const revenueData = [
  { month: 'Baisakh', revenue: 85000, expenses: 32000, profit: 53000 },
  { month: 'Jestha', revenue: 92000, expenses: 28000, profit: 64000 },
  { month: 'Ashadh', revenue: 78000, expenses: 45000, profit: 33000 },
  { month: 'Shrawan', revenue: 65000, expenses: 38000, profit: 27000 },
  { month: 'Bhadra', revenue: 110000, expenses: 42000, profit: 68000 },
  { month: 'Ashwin', revenue: 125000, expenses: 55000, profit: 70000 },
  { month: 'Kartik', revenue: 98000, expenses: 30000, profit: 68000 },
  { month: 'Mangsir', revenue: 88000, expenses: 35000, profit: 53000 },
  { month: 'Poush', revenue: 145000, expenses: 48000, profit: 97000 },
  { month: 'Magh', revenue: 130000, expenses: 40000, profit: 90000 },
  { month: 'Falgun', revenue: 105000, expenses: 36000, profit: 69000 },
  { month: 'Chaitra', revenue: 155000, expenses: 52000, profit: 103000 },
];

// ─── Recent Activity ────────────────────────────────────
export const recentActivity = [
  { id: '1', action: 'New request submitted', description: 'Sita Kumari Thapa requested 30 cft Sal Timber', time: '5 min ago', type: 'request' },
  { id: '2', action: 'Payment received', description: 'NPR 10,000 from Nirmala Devi Sharma', time: '1 hour ago', type: 'payment' },
  { id: '3', action: 'Request approved', description: 'Laxmi Maya Tamang - 80 cft Sal Timber', time: '2 hours ago', type: 'approval' },
  { id: '4', action: 'New member registered', description: 'Dipak Kumar Singh - Ward 5', time: '3 hours ago', type: 'member' },
  { id: '5', action: 'Fine collected', description: 'NPR 2,000 from Hari Prasad Adhikari', time: '5 hours ago', type: 'fine' },
  { id: '6', action: 'Letter issued', description: 'AGM Notice to all members', time: '1 day ago', type: 'letter' },
  { id: '7', action: 'Expense recorded', description: 'NPR 20,000 - Community fund contribution', time: '1 day ago', type: 'expense' },
];

// ─── Resource Distribution (pie chart) ──────────────────
export const resourceDistribution = [
  { name: 'Sal Timber', value: 38 },
  { name: 'Firewood', value: 28 },
  { name: 'Bamboo', value: 12 },
  { name: 'NTFP', value: 15 },
  { name: 'Others', value: 7 },
];

// ─── Ward Data ──────────────────────────────────────────
export const wardData = [
  { ward: 1, members: 42, active: 38 },
  { ward: 2, members: 35, active: 32 },
  { ward: 3, members: 28, active: 26 },
  { ward: 4, members: 31, active: 29 },
  { ward: 5, members: 20, active: 18 },
];

// ─── Monthly Financial Data ────────────────────────────
export const monthlyFinancialData = [
  { month: 'Baisakh', income: 85000, expense: 32000 },
  { month: 'Jestha', income: 92000, expense: 28000 },
  { month: 'Ashadh', income: 78000, expense: 45000 },
  { month: 'Shrawan', income: 65000, expense: 38000 },
  { month: 'Bhadra', income: 110000, expense: 42000 },
  { month: 'Ashwin', income: 125000, expense: 55000 },
  { month: 'Kartik', income: 98000, expense: 30000 },
  { month: 'Mangsir', income: 88000, expense: 35000 },
  { month: 'Poush', income: 145000, expense: 48000 },
  { month: 'Magh', income: 130000, expense: 40000 },
  { month: 'Falgun', income: 105000, expense: 36000 },
  { month: 'Chaitra', income: 155000, expense: 52000 },
];
