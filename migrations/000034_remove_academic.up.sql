-- Removes the academic and federated-learning structures that earlier phases
-- created. CampusCare AI is a community platform, so course/credential/learning
-- tables, the education-intervention tables and the federated-network tables are
-- not part of the product and must not linger in the schema.
--
-- Every statement is idempotent so the migration is a no-op on databases that
-- never had these tables, and safe to re-run on ones that do.
--
-- Events keep their own check-in tracking in attendance_records, which is keyed
-- off event_sessions and is deliberately left in place.

DROP TABLE IF EXISTS edu_digital_credentials CASCADE;
DROP TABLE IF EXISTS edu_credential_schemas CASCADE;
DROP TABLE IF EXISTS edu_transfer_credits CASCADE;
DROP TABLE IF EXISTS edu_mobility_profiles CASCADE;
DROP TABLE IF EXISTS edu_data_consents CASCADE;
DROP TABLE IF EXISTS edu_institutions CASCADE;

DROP TABLE IF EXISTS edu_assessments CASCADE;
DROP TABLE IF EXISTS edu_interventions CASCADE;
DROP TABLE IF EXISTS edu_learning_paths CASCADE;
DROP TABLE IF EXISTS edu_opportunities CASCADE;

DROP TABLE IF EXISTS network_learning_wallets CASCADE;
DROP TABLE IF EXISTS network_skill_equivalences CASCADE;
DROP TABLE IF EXISTS network_consents CASCADE;
DROP TABLE IF EXISTS network_data_contracts CASCADE;
DROP TABLE IF EXISTS network_nodes CASCADE;

-- student_profiles carried a course and university column, i.e. an academic
-- record, and was unused by the application.
DROP TABLE IF EXISTS student_profiles CASCADE;