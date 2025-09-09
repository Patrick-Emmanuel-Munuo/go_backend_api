const mysql = require("../configuration/mysql");
var helpers = require("../helper")
var mysql_backup = require('mysqldump');
const path = require('path');
var { sendMail } = require("../controllers/phone_email");

//database backup and email
async function backup() {
  try {
    // Generate filename and full path in public folder
    const time_now = Date.now();
    const file_name = `mysql_backup_${Math.floor(time_now / 1000)}.sql`;
    const file_path = path.join(__dirname, '../../public', file_name);
    // Get the dump as a string (safe with optional chaining)
    const dump = await mysql_backup?.({
      connection: {
        host: process.env.DB_HOST ?? 'localhost',
        user: process.env.DB_USER ?? 'root',
        password: process.env.DB_PASS ?? '',
        database: process.env.DB_NAME ?? 'vartrick',
      },
      dumpToFile: file_path,
    });
    // Email content
    const subject = "Database Backup File";
    const message = `Please find the attached backup file of the database created at ${new Date(time_now).toLocaleString()}.`;
    const html = `
      <p>Hello,</p>
      <p>Please find the database backup attached.</p>
      <p>Backup Time: ${new Date(time_now).toLocaleString()}</p>
      <p>Regards,<br/>BMH TSEMU TEAM</p>
    `;
    // Send email with attachment (safe with optional chaining)
    const emailResult = await sendMail?.({
      to: "patrickmunuo98@gmail.com",
      subject,
      message,
      html,
      attachments: [
        {
          filename: file_name,
          path: file_path,
        },
      ],
    });
    return {
      success: true,
      message: 'Backup created and sent via email.',
      email_response: emailResult?.success ?? false,
    };
  } catch (error) {
    console.log({ success: false, message: error?.message ?? "Unknown error" });
    return { success: false, message: error?.message ?? "Unknown error" };
  }
}
/* create and export function read data  */
async function read(options) {
  try {
    if (!options) {
      return { success: false, message: "Function 'options' parameter cannot be empty" };
    }
    // Destructure with defaults
    const {
      table = null,
      condition = null,
      select = null,
      or_condition = null
    } = options ?? {};
    if (!table || (!condition && !or_condition)) {
      return {
        success: false,
        message: "Table name and at least one of 'condition' or 'or_condition' is required"
      };
    }
    // Build SELECT fields
    const selectFields = select ? Object.keys(select ?? {}).join(", ") : "*";
    // Check for conditions
    const hasCondition = condition && Object.keys(condition ?? {}).length > 0;
    const hasOrCondition = or_condition && Object.keys(or_condition ?? {}).length > 0;
    
    // Build WHERE clause safely
    let whereClause = "1=1";
    if (hasCondition && hasOrCondition) {
      whereClause = `( ${helpers?.generateWhere?.(condition)} ) AND ( ${helpers?.generateWhere_or?.(or_condition)} )`;
    } else if (hasCondition) {
      whereClause = helpers?.generateWhere?.(condition);
    } else if (hasOrCondition) {
      whereClause = helpers?.generateWhere_or?.(or_condition);
    }
    // Prepare query safely
    const query = `SELECT ${selectFields} FROM ${mysql?.escapeId?.(table) ?? table} WHERE ${whereClause}`;
    // Execute query safely
    const [rows]  = await mysql.query(query);   
    if (!rows  || rows .length === 0) {
      return { success: false, message: "not found data" };
    } else { 
      return { success: true, message: rows  };

    } 
  } catch (error) {
    console.log("Read error: ", error?.message);
    return { success: false, message: error?.message ?? "Unknown error" };
  }
}
/* create and export function bulk read data  */
async function readBulk(options) {
  try {
    if (!options?.length || !Array.isArray(options)) {
      return { success: false, message: "Options must be a non-empty array" };
    }
    const data = [];
    const error = [];
    for (const opt of options) {
      const result = await read?.(opt) ?? { success: false, message: "Read function unavailable" };
      const entry = { table: opt?.table ?? "unknown", message: result?.message ?? "No message" };
      result?.success ? data.push(entry) : error.push(entry);
    }
    return {
      success: data?.length > 0 ? true : false,
      message: { data, error }
    };
  } catch (err) {
    return { success: false, message: err?.message ?? "Unknown error" };
  }
}
//list
async function list(options) {
  try {
    // Safe destructuring
    const table = options?.table;
    const select = options?.select;
    const page = parseInt(options?.page) || null;
    const limit = parseInt(options?.limit) || 10;
    const sort = options?.sort;
    if (!table) {
      return { success: false, message: 'table_name is required at body' };
    }
    // Build base query
    const selectFields = select ? Object.keys(select).join(', ') : '*';
    let baseQuery = `SELECT ${selectFields} FROM ${table}`;
    if (sort && Object.keys(sort).length > 0) {
      baseQuery += ` ORDER BY ${helpers.generateSort(sort)}`;
    }
    // Fetch all data first
    const [rows]  = await mysql?.query?.(baseQuery) ?? [];
    const totalRecords = rows?.length || 0;
    // If pagination requested
    if (page !== null) {
      const totalPages = Math.ceil(totalRecords / limit) || 1;
      const pages = Array.from({ length: totalPages }, (_, i) => i + 1);
      let currentPage = page;
      if (currentPage > totalPages) currentPage = totalPages;
      if (currentPage <= 0) currentPage = 1;
      const start = (currentPage - 1) * limit;
      const end = start + limit;
      const pageData = data.slice(start, end);
      return {
        success: true,
        message: {
          limit,
          totalRecords,
          totalPages,
          previous: currentPage > 1 ? currentPage - 1 : null,
          current: currentPage,
          next: currentPage < totalPages ? currentPage + 1 : null,
          pages,
          data: pageData
        }
      };
    }
    // If no pagination, return all data
    return { success: true, message: data };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}
// Delete a single record
async function delate(options) {
  try {
    const table = options?.table;
    const id = options?.id;
    if (!table || !id) {
      return { success: false, message: "Both 'table' and 'id' are required." };
    }
    const [result] = await mysql?.query?.(`DELETE FROM ${table} WHERE ${helpers.generateWhere({ id })}`);
    if (result?.affectedRows > 0) {
      return { success: true, message: "Data successfully deleted from the database." };
    }
    return { success: false, message: "No matching data found in the table." };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}
// Bulk delete records
async function delateBulk(options) {
  try {
    if (!Array.isArray(options) || options.length === 0) {
      return { success: false, message: "Options must be a non-empty array." };
    }
    const results = await Promise.allSettled(options.map(opt => deleteData(opt)));
    const dataDeleted = results
      .filter(r => r.status === "fulfilled" && r.value?.success)
      .map(r => r.value);
    const errorDeleted = results
      .filter(r => r.status === "fulfilled" && !r.value?.success)
      .map(r => r.value)
      .concat(
        results.filter(r => r.status === "rejected").map(r => ({ success: false, message: r.reason }))
      );
    if (dataDeleted?.length > 0 && errorDeleted?.length === 0) {
      return { success: true, message: dataDeleted };
    }
    if (dataDeleted?.length === 0 && errorDeleted?.length > 0) {
      return { success: false, message: errorDeleted };
    }
    return { success: false, message: [...errorDeleted, ...dataDeleted] };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}
// Create a single record
async function create(options) {
  try {
    const table = options?.table;
    const data = options?.data;
    if (!table || !data) {
      return { success: false, message: "Both 'table' and 'data' are required." };
    }
    const [result] = await mysql?.query?.(`INSERT INTO ${table} SET ?`, data);
    if (result?.affectedRows > 0) {
      return { success: true, message: { id: result.insertId, ...data } };
    }
    return { success: false, message: "Failed to insert data into the database." };
  } catch (error) {
    console.log({ success: false, message: error?.message });
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}
// Bulk create records using for loop (sequential)

async function createBulk(options) {
  try {
    const table = options?.table;
    const dataArray = options?.data;

    if (!table || !Array.isArray(dataArray)) {
      return { success: false, message: "'table' must be provided and 'data' must be an array." };
    }
    if (dataArray.length === 0) {
      return { success: false, message: "Data array cannot be empty." };
    }
    const dataCreate = [];
    const errorCreate = [];
    // Sequential insert using for loop
    for (const item of dataArray) {
      try {
        const result = await create({ table, data: item });
        if (result.success) {
          dataCreate.push(result);
        } else {
          errorCreate.push(result);
        }
      } catch (err) {
        errorCreate.push({ success: false, message: err?.message || "Unknown error occurred." });
      }
    }
    if (dataCreate.length > 0 && errorCreate.length === 0) {
      return { success: true, message: dataCreate };
    }
    if (dataCreate.length === 0 && errorCreate.length > 0) {
      return { success: false, message: errorCreate };
    }
    // Some succeeded, some failed
    return { success: false, message: [...errorCreate, ...dataCreate] };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}

/* create and export function count data  */
async function count(options) {
  try {
    const table = options?.table;
    const condition = options?.condition;
    if (!table || !condition || Object.keys(condition ?? {})?.length === 0) {
      return { success: false, message: "Both table_name and condition are required" };
    }
    const whereClause = helpers?.generateWhere?.(condition) ?? "1=1";
    const query = `SELECT COUNT(*) AS total FROM ${table} WHERE ${whereClause}`;
    const [rows] = await mysql?.query?.(query) ?? [];
    const total = rows?.[0]?.total ?? 0;
    return { success: true, message: total };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred" };
  }
}
/* create and export function to count bulk data using count() */
async function countBulk(options) {
  try {
    if (!Array.isArray(options) || options?.length === 0) {
      return { success: false, message: "Options must be a non-empty array" };
    }
    const results = [];
    for (const opt of options) {
      const table = opt?.table;
      const condition = opt?.condition;
      if (!table || !condition || Object.keys(condition ?? {})?.length === 0) {
        results.push({ table: table ?? "undefined", success: false, count: 0, message: "Table name and condition required" });
        continue;
      }
      const result = await count?.({ table, condition });
      results.push({
        table,
        success: result?.success ?? false,
        count: result?.success ? result?.message : 0,
        message: result?.success ? "Count retrieved" : result?.message
      });
    }
    //console.log({ success: true, message: results });
    return { success: true, message: results };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred" };
  }
}
//Search
async function search(options) {
  try {
    // Destructure options safely with defaults
    const {
      table = null,
      condition = null,
      select = null,
      or_condition = null
    } = options ?? {};
    // Validate required parameters
       if (!table || (!condition && !or_condition)) {
      return {
        success: false,
        message: "Table name and at least one of 'condition' or 'or_condition' is required"
      };
    }
    // Determine SELECT fields
    const selectFields = select ? Object.keys(select ?? {}).join(", ") : "*";
    // Build WHERE clause using helpers
    const hasCondition = condition && Object.keys(condition ?? {}).length > 0;
    const hasOrCondition = or_condition && Object.keys(or_condition ?? {}).length > 0;
    
    // Build WHERE clause safely
    let whereClause = "1=1";
    if (hasCondition && hasOrCondition) {
      whereClause = `( ${helpers?.generateLike?.(condition)} ) AND ( ${helpers?.generateWhere_or?.(or_condition)} )`;
    } else if (hasCondition) {
      whereClause = helpers?.generateLike?.(condition);
    } else if (hasOrCondition) {
      whereClause = helpers?.generateWhere_or?.(or_condition);
    }
    // Construct query
    const query = `SELECT ${selectFields} FROM ${table} WHERE ${whereClause}`;
    // Execute query safely
    const [rows] = await mysql?.query?.(query);
    // Return response based on result
    if (!rows || rows.length === 0) {
      return { success: false, message: "No matching data found." };
    }
    return { success: true, message: rows };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}
//Search between
async function search_between(options) {
  try {
    // Safe destructuring with defaults
    const {
      table,
      column,
      start,
      end,
      select = null,
      condition = {}
    } = options ?? {};
    console.log({ options });
    // Validate required parameters
    if (!table || !column || start == null || end == null) {
      return {
        success: false,
        message: "Parameters 'table', 'column', 'start', and 'end' are required."
      };
    }
    // Determine SELECT fields
    const selectFields = select ? Object.keys(select).join(", ") : "*";
    // Build WHERE clause for extra conditions
    let whereClause = "1=1";
    const params = [start, end]; // params for BETWEEN
    if (condition && Object.keys(condition).length > 0) {
      const conditionStrings = [];
      Object.entries(condition).forEach(([key, value]) => {
        conditionStrings.push(`${key} = ?`);
        params.push(value);
      });
      whereClause += " AND " + conditionStrings.join(" AND ");
    }
    // Construct query
    const query = `SELECT ${selectFields} FROM ${table} WHERE ${column} BETWEEN ? AND ? AND ${whereClause}`;
    console.log({ query });
    // Execute query safely
    const [rows] = await mysql?.query?.(query, params);
    // Return response based on result
    console.log({ rows });
    if (!rows || rows.length === 0) {
      return { success: false, message: "No records found in the given range." };
    }
    return { success: true, message: rows };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred." };
  }
}
//create data
async function update(options) {
  try {
    // Destructure options with defaults
    const {
      table = null,
      data = null,
      condition = null,
      encript = null
    } = options ?? {};
    // Validate required parameters
    if (!table || !data || !condition) {
      return {
        success: false,
        message: "Table name, updated data, and condition are required."
      };
    }
    // Apply encryption if specified
    const finalData = encript ? helpers.generateCreateEncrypt(data, encript) : data;
    // Build SQL query
    const query = `
      UPDATE ${table} 
      SET ${helpers.updateSet(finalData)} 
      WHERE ${helpers.generateWhere(condition)}
    `;
    // Execute query
    const [result] = await mysql?.query?.(query);
    console.log({ result });
    // Check result and return appropriate response
    // Handle response
    if (result?.affectedRows > 0 && result?.changedRows > 0) {
      return { success: true, message: "Data updated successfully." };
    } else if (result?.affectedRows > 0 && result?.changedRows === 0) {
      return { success: false, message: "No changes made. Data is identical to existing records." };
    } else {
      return { success: false, message: "Failed to update data." };
    }
  } catch (error) {
    console.log({ success: false, message: error });
    return { success: false, message: error?.message ?? "Unknown error" };
  }
}
/* create and export function updateBulk data  */
async function updateBulk(options) {
  try {
    if (!options) {
      return { success: false, message: "Function 'options' parameter is required" };
    }
    if (!Array.isArray(options)) {
      return { success: false, message: "Function 'options' must be an array" };
    }
    if (options.length === 0) {
      return { success: false, message: "Options array cannot be empty" };
    }
    const dataUpdate = [];
    const errorUpdate = [];
    for (const opt of options) {
      const result = await update(opt);
      if (result?.success) {
        dataUpdate.push(result);
      } else {
        errorUpdate.push(result);
      }
    }
    const hasSuccess = dataUpdate?.length > 0;
    const hasError = errorUpdate?.length > 0;
    if (hasSuccess && !hasError) {
      return { success: true, message: dataUpdate };
    }
    if (!hasSuccess && hasError) {
      return { success: false, message: errorUpdate };
    }
    if (hasSuccess && hasError) {
      // Partial success
      return { success: true, message: [...errorUpdate, ...dataUpdate] };
    }
    return { success: false, message: "No data processed" };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error occurred" };
  }
}
/* mysql querry */
async function query(options) {
  try {
    const sql = options?.query;
    if (!sql) {
      return { success: false, message: "query is required at body in request mysql query" };
    }
    const result = await mysql?.query?.(sql);
    if (!result?.length) {
      return { success: false, message: "No data found" };
    }
    return { success: true, message: result };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error" };
  }
}
/* databases  */
async function databaseHandler(options) {
  try {
    if (!options) {
      return { success: false, message: "Function options parameter cannot be empty" };
    }
    const { action, database, table, newtable, column, columns } = options ?? {};
    if (!action) {
      return { success: false, message: "Action is required" };
    }
    let query = "";
    let successMessage = "";
    switch (action) {
      // DATABASE
      case "create_db":
        if (!database) return { success: false, message: "database_name is required" };
        query = `CREATE DATABASE ${mysql.escapeId(database)}`;
        successMessage = `Database '${database}' created successfully`;
        break;
      case "delete_db":
        if (!database) return { success: false, message: "database_name is required" };
        query = `DROP DATABASE IF EXISTS ${mysql.escapeId(database)}`;
        successMessage = `Database '${database}' deleted successfully`;
        break;
      case "update_db":
        if (!database || !newtable) return { success: false, message: "Both old and new database_name are required" };
        query = `ALTER DATABASE ${mysql.escapeId(database)} UPGRADE DATA DIRECTORY NAME`; // MySQL doesn't directly support rename db; usually requires dump+import
        successMessage = `Database '${database}' updated successfully`;
        break;
      // TABLE
      case "create_table":
        if (!table) return { success: false, message: "table_name is required" };
        if (!columns || !Array.isArray(columns) || columns.length === 0) {
          return { success: false, message: "columns array is required for creating table" };
        }
        const columnsDef = columns.map(col => `${mysql.escapeId(col.name)} ${col.type}`).join(", ");
        query = `CREATE TABLE ${mysql.escapeId(table)} (${columnsDef})`;
        successMessage = `Table '${table}' created successfully`;
        break;
      case "update_table":
        if (!table || !newtable) return { success: false, message: "Both table_name and new table_name are required" };
        query = `ALTER TABLE ${mysql.escapeId(table)} RENAME TO ${mysql.escapeId(newtable)}`;
        successMessage = `Table renamed from '${table}' to '${newtable}' successfully`;
        break;
      case "delete_table":
        if (!table) return { success: false, message: "table_name is required" };
        query = `DROP TABLE IF EXISTS ${mysql.escapeId(table)}`;
        successMessage = `Table '${table}' deleted successfully`;
        break;
      // COLUMN
      case "create_column":
        if (!table || !column) return { success: false, message: "Both table_name and column definition are required" };
        query = `ALTER TABLE ${mysql.escapeId(table)} ADD ${column}`; // column should include name and type e.g., `name VARCHAR(50)`
        successMessage = `Column added to table '${table}' successfully`;
        break;
      case "update_column":
        if (!table || !column || !options.newColumn) return { success: false, message: "table_name, column and newColumn are required" };
        query = `ALTER TABLE ${mysql.escapeId(table)} CHANGE ${column} ${options.newColumn}`; // newColumn should include new name and type
        successMessage = `Column '${column}' updated in table '${table}' successfully`;
        break;
      case "delete_column":
        if (!table || !column) return { success: false, message: "Both table_name and column are required" };
        query = `ALTER TABLE ${mysql.escapeId(table)} DROP COLUMN ${mysql.escapeId(column)}`;
        successMessage = `Column '${column}' removed from table '${table}' successfully`;
        break;
      default:
        return { success: false, message: `Unknown action: '${action}'` };
    }
    // Execute query once
    await mysql?.query?.(query);
    return { success: true, message: successMessage };
  } catch (error) {
    return { success: false, message: error?.message ?? "Unknown error" };
  }
}
module.exports = {
  query,
  backup,
  updateBulk,
  update,
  search,
  search_between,
  count,
  countBulk,
  create,
  createBulk,
  delate,
  delateBulk,
  listAll:list,
  list,
  read,
  readBulk,
  databaseHandler
};