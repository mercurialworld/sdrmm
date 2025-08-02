#[derive(Debug)]
pub struct QueueStatus {
    pub timestamp: i32,
    pub open: bool,
}

#[derive(Debug)]
pub struct SessionRequests {
    pub user: String,
    pub requests: i32,
}
