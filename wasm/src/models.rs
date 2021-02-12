use serde::{Deserialize, Serialize};

/// Use the InputModel to model an incoming message
#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
pub struct InputModel {
    pub foo: i32,
}

/// Example trait for processing inputs
pub trait Process {
    fn process(&self) -> serde_json::Value;
}

/// Implement this method to transform incoming messages
impl Process for InputModel {
    fn process(&self) -> serde_json::Value {
        let output_data = InputModel { foo: self.foo + 1 };
        match serde_json::to_value(output_data) {
            Ok(parsed_json) => parsed_json,
            Err(err) => {
                panic!(err);
            }
        }
    }
}
