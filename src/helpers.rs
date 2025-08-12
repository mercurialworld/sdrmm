use num::Num;

#[expect(unused)]
/// Checks if you need to ignore `config_val`.
/// If that is false, checks if `to_compare` is greater than `config_val`.
///
/// Returns true if `config_val` is 0 or `to_compare` > `config_val`.
pub fn ignore_or_gt<T: Num + PartialOrd + Clone>(to_compare: T, config_val: T) -> bool {
    ignore_config(config_val.clone()) || to_compare > config_val
}

/// Checks if you need to ignore `config_val`.
/// If that is false, checks if `to_compare` is less than `config_val`.
///
/// Returns true if `config_val` is 0 or `to_compare` < `config_val`.
pub fn ignore_or_lt<T: Num + PartialOrd + Clone>(to_compare: T, config_val: T) -> bool {
    ignore_config(config_val.clone()) || to_compare < config_val
}

/// Checks if you need to ignore `config_val`.
/// If that is false, checks if `to_compare` is greater than or equal to `config_val`.
///
/// Returns true if `config_val` is 0 or `to_compare` >= `config_val`.
pub fn ignore_or_geq<T: Num + PartialOrd + Clone>(config_val: T, to_compare: T) -> bool {
    ignore_config(config_val.clone()) || to_compare >= config_val
}

/// Checks if you need to ignore `config_val`.
/// If that is false, checks if `to_compare` is less than or equal to `config_val`.
///
/// Returns true if `config_val` is 0 or `to_compare` <= `config_val`.
pub fn ignore_or_leq<T: Num + PartialOrd + Clone>(config_val: T, to_compare: T) -> bool {
    ignore_config(config_val.clone()) || to_compare <= config_val
}

/// Checks if you need to ignore `config_val`.
/// If that is false, checks if any `T` in the `to_compare` vector is greater than `config_val`.
///
/// Returns true if there's at least one value that meets requirements, or the setting is 0.
pub fn ignore_or_geq_vec<T: Num + PartialOrd + Clone>(to_compare: &Vec<T>, config_val: T) -> bool {
    if ignore_config(config_val.clone()) {
        return true;
    }

    let mut one_diff_meets_criteria = false;

    for diff_val in to_compare {
        if *diff_val >= config_val {
            one_diff_meets_criteria = true;
        }
    }

    one_diff_meets_criteria
}

/// Checks if you need to ignore `config_val`.
/// If that is false, checks if any `T` in the `to_compare` vector is less than `config_val`.
///
/// Returns true if there's at least one value that meets requirements, or the setting is 0.
pub fn ignore_or_leq_vec<T: Num + PartialOrd + Clone>(to_compare: &Vec<T>, config_val: T) -> bool {
    if ignore_config(config_val.clone()) {
        return true;
    }

    let mut one_diff_meets_criteria = false;

    for diff_val in to_compare {
        if *diff_val <= config_val {
            one_diff_meets_criteria = true;
        }
    }

    one_diff_meets_criteria
}

/// Checks if at least one string in one Vec of strings
/// is in another Vec of strings.
///
/// Returns true if a string of one string Vec is found in another string Vec.
pub fn match_in_two_vecs(to_find: Vec<String>, find_in: Vec<String>) -> bool {
    for item in to_find {
        if let Some(_) = find_in.iter().find(|&s| item.eq(s)) {
            return true;
        }
    }

    false
}

/// Checks if a value is zero.
///
/// Returns true if the value is 0, false otherwise.
pub fn ignore_config<T: Num>(val: T) -> bool {
    val.is_zero()
}
