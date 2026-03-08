pub fn add(a: i32, b: i32) -> i32 {
    a + b
}

pub fn is_positive(n: i32) -> bool {
    if n > 0 {
        return true;
    }
    false
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add() {
        assert_eq!(add(2, 3), 5);
    }

    #[test]
    fn test_is_positive() {
        assert!(is_positive(1));
        assert!(!is_positive(-1));
    }
}
