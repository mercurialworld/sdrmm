use chrono::{DateTime, Utc};
use rusqlite::Connection;
use thiserror::Error;

use crate::database::schema::QueueStatus;

pub(crate) mod schema;

#[derive(Debug, Error)]
pub enum DatabaseError {
    #[error("RuSQLite error")]
    SqliteError(#[from] rusqlite::Error),
    #[error("User not found")]
    NotFound(i32),
}

type DBResult<T> = Result<T, DatabaseError>;

pub struct Database {
    conn: Connection,
}

impl Database {
    pub fn from_file(filename: &str) -> DBResult<Self> {
        let conn = Connection::open(filename)?;

        Ok(Self { conn })
    }

    pub fn init_db(&self) -> DBResult<()> {
        self.conn.execute(
            "CREATE TABLE IF NOT EXISTS QueueStatus(
            timestamp INTEGER PRIMARY KEY,
            open      INTEGER
        )",
            (),
        )?;

        self.conn.execute(
            "CREATE TABLE IF NOT EXISTS SessionRequests(
            user      TEXT PRIMARY KEY,
            requests  INTEGER
        )",
            (),
        )?;

        Ok(())
    }

    pub fn new_session(&self, timestamp: DateTime<Utc>, open: bool) -> DBResult<()> {
        self.conn.execute(
            "INSERT INTO QueueStatus (timestamp, open) VALUES (?1, ?2)",
            (&timestamp.timestamp(), &open),
        )?;

        Ok(())
    }

    pub fn get_queue_status(&self) -> DBResult<bool> {
        let mut query = self.conn.prepare(
            "
            SELECT timestamp, open
            FROM QueueStatus 
            ORDER BY timestamp DESC 
            LIMIT 1
            ",
        )?;

        let mut result = query.query_map([], |row| {
            Ok(QueueStatus {
                timestamp: row.get(0)?,
                open: row.get(1)?,
            })
        })?;

        // [FIXME] surely there's something better
        Ok(result.next().unwrap().unwrap().open)
    }

    pub fn set_queue_status(&self, open: bool) -> DBResult<()> {
        self.conn.execute(
            "
            WITH CurrentSession AS (
                SELECT timestamp, open 
                FROM QueueStatus 
                ORDER BY timestamp DESC 
                LIMIT 1
            )
            UPDATE QueueStatus 
            SET open=?1 
            WHERE timestamp=(SELECT timestamp FROM CurrentSession)",
            (&open,),
        )?;

        Ok(())
    }

    pub fn get_user_requests(&self, user: &str) -> DBResult<i32> {
        let mut query = self
            .conn
            .prepare("SELECT requests FROM SessionRequests WHERE user=?1")?;

        let mut rows = query.query([user])?;
        let mut results: Vec<i32> = Vec::new();
        while let Some(row) = rows.next()? {
            results.push(row.get(0)?);
        }

        // what is this
        results.get(0).ok_or(DatabaseError::NotFound(-1)).copied()
    }

    pub fn add_user_requests(&self, user: &str) -> DBResult<()> {
        self.conn.execute(
            "INSERT INTO SessionRequests(user, requests) VALUES ?1, 0",
            (user,),
        )?;

        Ok(())
    }

    pub fn set_user_requests(&self, user: &str, requests: i32) -> DBResult<()> {
        self.conn.execute(
            "UPDATE SessionRequests SET requests=?1 WHERE user=?2",
            (requests, user),
        )?;

        Ok(())
    }

    pub fn clear_user_requests(&self) -> DBResult<()> {
        self.conn.execute("DELETE FROM SessionRequests", ())?;

        Ok(())
    }
}
