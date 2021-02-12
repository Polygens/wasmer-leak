use crate::models::Process;
use core::slice;
use core::str;

mod models;

/// Export a function named `transform`. This can be called
/// from the embedder!
#[no_mangle]
pub extern "C" fn transform(ptr: i32, len: i32) -> i64 {
    let slice = unsafe { slice::from_raw_parts(ptr as _, len as _).to_vec() };
    let string_from_host = str::from_utf8(&slice).unwrap();

    let output = match serde_json::from_str::<models::InputModel>(string_from_host) {
        Ok(input_model) => {
            let result = input_model.process();
            serde_json::to_string(&result).unwrap()
        }
        Err(err) => {
            panic!(err);
        }
    };

    cantor_pairing(output.as_ptr() as i32, output.len() as i32)
}

fn cantor_pairing(k1: i32, k2: i32) -> i64 {
    (k1 as i64 + k2 as i64) * (k1 as i64 + k2 as i64 + 1) / 2 + k2 as i64
}
