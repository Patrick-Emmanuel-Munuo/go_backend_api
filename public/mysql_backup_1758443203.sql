-- Table structure for `departments`
CREATE TABLE `departments` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `delated` tinyint(1) NOT NULL DEFAULT 0,
  `department_name` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `status` varchar(20) DEFAULT 'active',
  `location` varchar(150) DEFAULT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `hod` int(11) DEFAULT NULL,
  `assist_hod` int(11) DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_date` datetime DEFAULT current_timestamp(),
  `updated_by` int(11) DEFAULT NULL,
  `updated_date` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `department_code` varchar(29) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `hod` (`hod`),
  KEY `assist_hod` (`assist_hod`),
  KEY `created_by` (`created_by`),
  KEY `updated_by` (`updated_by`),
  CONSTRAINT `departments_ibfk_1` FOREIGN KEY (`hod`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `departments_ibfk_2` FOREIGN KEY (`assist_hod`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `departments_ibfk_3` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `departments_ibfk_4` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8mb4;

INSERT INTO `departments` (id, delated, department_name, description, status, location, phone, hod, assist_hod, created_by, created_date, updated_by, updated_date, department_code) VALUES ('2', '0', '[69 108 101 99 116 114 105 99 97 108 32 69 110 103]', '[69 110 103 105 110 101 101 114 105 110 103 32 119 111 114 107 115 32 114 101 108 97 116 101 100 32 116 111 32 101 108 101 99 116 114 105 99 97 108 32 97 110 100 32 65 67 32 115 121 115 116 101 109 115]', '[97 99 116 105 118 101]', '[82 79 79 77 32 50 48 48]', '[43 50 53 53 54 50 53 52 52 57 50 57 53]', '15', NULL, '15', '2024-08-16 12:18:36 +0300 EAT', '15', '2025-09-10 02:10:04 +0300 EAT', '[66 77 72 48 48 49]');
INSERT INTO `departments` (id, delated, department_name, description, status, location, phone, hod, assist_hod, created_by, created_date, updated_by, updated_date, department_code) VALUES ('4', '0', '[66 105 111 109 101 100 105 99 97 108 32 69 110 103]', '[77 97 105 110 116 101 110 97 110 99 101 32 97 110 100 32 109 97 110 97 103 101 109 101 110 116 32 111 102 32 98 105 111 109 101 100 105 99 97 108 32 101 113 117 105 112 109 101 110 116]', '[97 99 116 105 118 101]', '[82 79 79 77 32 50 48 51]', '[49 50 51 52]', NULL, NULL, '15', '2024-08-16 12:18:36 +0300 EAT', '15', '2025-09-10 02:24:32 +0300 EAT', '[66 77 72 48 48 50]');
INSERT INTO `departments` (id, delated, department_name, description, status, location, phone, hod, assist_hod, created_by, created_date, updated_by, updated_date, department_code) VALUES ('67', '0', '[80 108 117 109 98 105 110 103]', '[80 108 117 109 98 105 110 103 44 32 115 97 110 105 116 97 116 105 111 110 44 32 97 110 100 32 119 97 116 101 114 32 115 121 115 116 101 109 32 109 97 110 97 103 101 109 101 110 116]', '[97 99 116 105 118 101]', '[82 79 79 77 32 50 48 52]', '[43 50 53 53 54 50 53 52 52 57 50 57 53]', '15', '15', '15', '2024-08-16 12:18:36 +0300 EAT', '15', '2025-09-10 10:09:10 +0300 EAT', '[66 77 72 48 48 51]');
INSERT INTO `departments` (id, delated, department_name, description, status, location, phone, hod, assist_hod, created_by, created_date, updated_by, updated_date, department_code) VALUES ('68', '0', '[67 105 118 105 108 32 97 110 100 32 69 110 103]', '[67 105 118 105 108 32 119 111 114 107 115 44 32 99 111 110 115 116 114 117 99 116 105 111 110 44 32 97 110 100 32 98 117 105 108 100 105 110 103 32 109 97 105 110 116 101 110 97 110 99 101]', '[97 99 116 105 118 101]', '[66 97 115 101 109 101 110 116 32 66 108 111 99 107 32 50]', '[49 50 51 52]', '15', '15', '15', '2025-05-12 09:32:10 +0300 EAT', '15', '2025-09-10 10:09:21 +0300 EAT', '[66 77 72 48 48 52]');
INSERT INTO `departments` (id, delated, department_name, description, status, location, phone, hod, assist_hod, created_by, created_date, updated_by, updated_date, department_code) VALUES ('70', '0', '[77 101 99 104 97 110 105 99 97 108]', '[66 111 105 108 101 114 32 111 112 101 114 97 116 105 111 110 115 44 32 119 101 108 100 105 110 103 44 32 97 110 100 32 114 101 108 97 116 101 100 32 109 101 99 104 97 110 105 99 97 108 32 119 111 114 107 115]', '[97 99 116 105 118 101]', '[87 111 114 107 115 104 111 112 32 65 114 101 97]', '[48 55 49 50 51 52 53 54 55 56]', '15', '15', '15', '2025-05-17 03:32:58 +0300 EAT', '15', '2025-09-10 10:09:29 +0300 EAT', '[66 77 72 48 48 53]');
INSERT INTO `departments` (id, delated, department_name, description, status, location, phone, hod, assist_hod, created_by, created_date, updated_by, updated_date, department_code) VALUES ('71', '0', '[80 108 117 109 98 105 110 103 121 121]', '[80 108 117 109 98 105 110 103 44 32 115 97 110 105 116 97 116 105 111 110 44 32 97 110 100 32 119 97 116 101 114 32 115 121 115 116 101 109 32 109 97 110 97 103 101 109 101 110 116]', '[97 99 116 105 118 101]', '[103 103 103 103 103 103]', '[53 53 53 53 53 53 53 53 53 53 53 53 53 53 53 53]', NULL, NULL, '15', '2025-09-10 10:23:24 +0300 EAT', NULL, '2025-09-10 10:23:24 +0300 EAT', '[66 77 72 48 48 54]');

-- Table structure for `inspections`
CREATE TABLE `inspections` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `delated` tinyint(1) NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `inspection_department_id` int(11) NOT NULL DEFAULT 0,
  `inspected_name` varchar(200) NOT NULL,
  `inspection_date` datetime NOT NULL DEFAULT current_timestamp(),
  `inspection_status` varchar(100) NOT NULL,
  `recommendations` text DEFAULT NULL,
  `comments` text DEFAULT NULL,
  `technical_finding` text DEFAULT NULL,
  `reported_department` int(11) DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_date` datetime DEFAULT current_timestamp(),
  `updated_by` int(11) DEFAULT NULL,
  `updated_date` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `inspection_department_id` (`inspection_department_id`),
  KEY `reported_department` (`reported_department`),
  KEY `created_by` (`created_by`),
  KEY `updated_by` (`updated_by`),
  CONSTRAINT `inspections_ibfk_1` FOREIGN KEY (`inspection_department_id`) REFERENCES `departments` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `inspections_ibfk_2` FOREIGN KEY (`reported_department`) REFERENCES `departments` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `inspections_ibfk_3` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `inspections_ibfk_4` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4;

INSERT INTO `inspections` (id, delated, status, inspection_department_id, inspected_name, inspection_date, inspection_status, recommendations, comments, technical_finding, reported_department, created_by, created_date, updated_by, updated_date) VALUES ('1', '0', '[97 99 116 105 118 101]', '2', '[103 111 111 111 100]', '2025-09-09 21:54:59 +0300 EAT', '[79 75]', '[103 111 111 111 100]', '[103 111 111 111 100]', '[103 111 111 111 100]', '2', '15', '2025-09-09 21:54:59 +0300 EAT', '15', '2025-09-10 00:34:01 +0300 EAT');
INSERT INTO `inspections` (id, delated, status, inspection_department_id, inspected_name, inspection_date, inspection_status, recommendations, comments, technical_finding, reported_department, created_by, created_date, updated_by, updated_date) VALUES ('2', '0', '[97 99 116 105 118 101]', '2', '[103 111 111 100 32 108 105 103 104 116]', '2025-09-10 00:33:47 +0300 EAT', '[79 75]', '[103 111 111 100 32 108 105 103 104 116]', '[103 111 111 100 32 108 105 103 104 116]', '[103 111 111 100 32 108 105 103 104 116]', '2', '15', '2025-09-10 00:33:47 +0300 EAT', '15', '2025-09-10 00:34:40 +0300 EAT');
INSERT INTO `inspections` (id, delated, status, inspection_department_id, inspected_name, inspection_date, inspection_status, recommendations, comments, technical_finding, reported_department, created_by, created_date, updated_by, updated_date) VALUES ('3', '0', '[97 99 116 105 118 101]', '68', '[109 97 99 104 105 110 101 32 120 114 97 121]', '2025-09-10 14:42:56 +0300 EAT', '[79 75]', '[103 111 111 100]', NULL, '[105 107 111 32 103 111 111 100 32]', '2', '15', '2025-09-10 14:42:56 +0300 EAT', NULL, '2025-09-10 14:42:56 +0300 EAT');

-- Table structure for `inventory`
CREATE TABLE `inventory` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `item_name` varchar(255) DEFAULT NULL,
  `descriptions` text DEFAULT NULL,
  `user_department` varchar(255) DEFAULT NULL,
  `location` varchar(255) DEFAULT NULL,
  `specifications` text DEFAULT NULL,
  `installed_date` varchar(50) DEFAULT NULL,
  `active` float(10,3) DEFAULT 1.000,
  `inactive` float(10,3) DEFAULT 0.000,
  `last_ppm` varchar(50) DEFAULT NULL,
  `next_ppm` varchar(50) DEFAULT NULL,
  `status` varchar(50) DEFAULT 'Operational',
  `purchase_date` varchar(30) DEFAULT NULL,
  `created_date` varchar(50) DEFAULT NULL,
  `created_by` varchar(100) DEFAULT NULL,
  `aprooved` tinyint(1) DEFAULT NULL,
  `delated` tinyint(50) DEFAULT NULL,
  `comment` varchar(110) DEFAULT NULL,
  `created_department` varchar(50) DEFAULT NULL,
  `updated_date` varchar(50) DEFAULT NULL,
  `updated_by` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4;

INSERT INTO `inventory` (id, item_name, descriptions, user_department, location, specifications, installed_date, active, inactive, last_ppm, next_ppm, status, purchase_date, created_date, created_by, aprooved, delated, comment, created_department, updated_date, updated_by) VALUES ('1', '[76 69 68 32 49 50 87]', '[108 101 100 32 99 101 105 108 105 110 103 32 108 97 109 112]', '[54 54 98 55 98 50 99 98 50 102 98 54 57 51 100 49 98 100 56 52 98 52 49 53]', '[69 78 84]', '[49 55 53 81 32 50 50 48 86 97 99]', '[50 48 50 53 45 48 53 45 48 54]', '10', '0', '[50 48 50 53 45 48 53 45 50 51]', '[50 48 50 53 45 48 53 45 50 49]', '[97 99 116 105 118 101]', '[50 48 50 53 45 48 53 45 50 48]', '[]', '[]', '0', '0', '[]', '[54 54 98 55 98 50 99 98 50 102 98 54 57 51 100 49 98 100 56 52 98 52 49 53]', '[49 57 45 48 56 45 50 48 50 53 84 49 51 58 51 51 58 48 55]', '[54 54 97 99 51 99 50 50 53 53 99 97 98 53 49 53 55 54 99 100 52 100 49 48 50]');
INSERT INTO `inventory` (id, item_name, descriptions, user_department, location, specifications, installed_date, active, inactive, last_ppm, next_ppm, status, purchase_date, created_date, created_by, aprooved, delated, comment, created_department, updated_date, updated_by) VALUES ('2', '[102 102 102 102 102]', '[114 114]', '[54 54 98 53 50 53 55 53 49 100 53 51 57 49 98 56 48 50 99 51 52 102 57 99]', '[102 102 102 102 102 102]', '[102]', '[50 48 50 53 45 48 56 45 49 50]', '55', '5', '[50 48 50 53 45 48 56 45 50 49]', '[50 48 50 53 45 48 56 45 50 48]', '[97 99 116 105 118 101]', '[50 48 50 53 45 48 56 45 49 52]', '[49 53 45 48 56 45 50 48 50 53 84 49 54 58 48 55 58 51 52]', '[54 54 97 99 51 99 50 50 53 53 99 97 98 53 49 53 55 54 99 100 52 100 49 48 50]', NULL, '0', NULL, '[54 54 98 55 98 50 99 98 50 102 98 54 57 51 100 49 98 100 56 52 98 52 49 53]', '[49 57 45 48 56 45 50 48 50 53 84 49 51 58 51 51 58 48 55]', '[54 54 97 99 51 99 50 50 53 53 99 97 98 53 49 53 55 54 99 100 52 100 49 48 50]');

-- Table structure for `jobs`
CREATE TABLE `jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `delated` tinyint(1) NOT NULL DEFAULT 0,
  `requested_department` int(11) DEFAULT NULL,
  `descriptions` text NOT NULL,
  `technical_specifications` text DEFAULT NULL,
  `materials_required` text DEFAULT NULL,
  `material_used` text DEFAULT NULL,
  `instrument` varchar(100) DEFAULT NULL,
  `location` varchar(100) DEFAULT NULL,
  `status` text DEFAULT NULL,
  `progress` decimal(5,2) NOT NULL DEFAULT 0.00,
  `emergency` tinyint(1) NOT NULL DEFAULT 0,
  `comment` text DEFAULT NULL,
  `customer_comment` text DEFAULT NULL,
  `customer_approve` tinyint(1) NOT NULL DEFAULT 0,
  `supervisor_comment` text DEFAULT NULL,
  `supervisor_approve` tinyint(1) NOT NULL DEFAULT 0,
  `attempted` tinyint(1) NOT NULL DEFAULT 0,
  `assigned_department` int(11) DEFAULT NULL,
  `assigned_to` int(11) DEFAULT NULL,
  `assigned_by` int(11) DEFAULT NULL,
  `assigned_date` datetime DEFAULT NULL,
  `assigned_comment` text DEFAULT NULL,
  `assigned_report_time` datetime DEFAULT NULL,
  `assigned_referred` int(11) DEFAULT NULL,
  `attempt_by` int(11) DEFAULT NULL,
  `attempt_date` datetime DEFAULT NULL,
  `attempt_comment` text DEFAULT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_date` datetime DEFAULT current_timestamp(),
  `updated_by` int(11) DEFAULT NULL,
  `updated_date` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `printed` tinyint(1) NOT NULL DEFAULT 0,
  `printed_date` datetime DEFAULT NULL,
  `printed_by` int(11) DEFAULT NULL,
  `year` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `assigned_to` (`assigned_to`),
  KEY `assigned_referred` (`assigned_referred`),
  KEY `updated_by` (`updated_by`),
  KEY `printed_by` (`printed_by`),
  KEY `created_by` (`created_by`),
  KEY `assigned_by` (`assigned_by`),
  KEY `attempt_by` (`attempt_by`),
  KEY `assigned_department` (`assigned_department`),
  KEY `requested_department` (`requested_department`),
  CONSTRAINT `jobs_ibfk_1` FOREIGN KEY (`assigned_to`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_2` FOREIGN KEY (`assigned_referred`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_3` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_4` FOREIGN KEY (`printed_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_5` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_6` FOREIGN KEY (`assigned_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_7` FOREIGN KEY (`attempt_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_8` FOREIGN KEY (`assigned_department`) REFERENCES `departments` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `jobs_ibfk_9` FOREIGN KEY (`requested_department`) REFERENCES `departments` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8mb4;

INSERT INTO `jobs` (id, delated, requested_department, descriptions, technical_specifications, materials_required, material_used, instrument, location, status, progress, emergency, comment, customer_comment, customer_approve, supervisor_comment, supervisor_approve, attempted, assigned_department, assigned_to, assigned_by, assigned_date, assigned_comment, assigned_report_time, assigned_referred, attempt_by, attempt_date, attempt_comment, created_by, created_date, updated_by, updated_date, printed, printed_date, printed_by, year) VALUES ('18', '0', '2', '[115 115 115 115 115 115 115]', NULL, NULL, NULL, '[115 115 115 115 115 115 115 115 115 115 115 115 115]', '[115 115 115 115 115 115 115]', '[]', '[48 46 48 48]', '0', '[116 97 115 107 32 114 101 45 114 101 113 117 101 115 116 101 100 32 119 97 105 116 32 102 111 114 32 97 115 115 105 103 110 109 101 110 116]', NULL, '0', NULL, '0', '0', '2', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '15', '2025-09-01 05:21:22 +0300 EAT', '15', '2025-09-10 14:34:36 +0300 EAT', '0', NULL, NULL, NULL);
INSERT INTO `jobs` (id, delated, requested_department, descriptions, technical_specifications, materials_required, material_used, instrument, location, status, progress, emergency, comment, customer_comment, customer_approve, supervisor_comment, supervisor_approve, attempted, assigned_department, assigned_to, assigned_by, assigned_date, assigned_comment, assigned_report_time, assigned_referred, attempt_by, attempt_date, attempt_comment, created_by, created_date, updated_by, updated_date, printed, printed_date, printed_by, year) VALUES ('19', '0', '2', '[115 115 115 115 115 115 115 115 115 115 115 115 115]', '[103 106 103 106 103 106 103 106 103]', NULL, NULL, '[101 101 101 101 101 101 101 101 101]', '[104 104 104 104 104 104 104]', '[99 111 109 112 108 101 116 101 100]', '[48 46 48 48]', '1', '[84 97 115 107 32 97 115 115 105 103 110 101 100 44 32 119 97 105 116 32 102 111 114 32 97 116 116 101 109 112 116 105 110 103]', '[103 111 111 100]', '1', '[103 111 111 100]', '1', '1', '2', '15', '15', NULL, '[80 114 111 99 101 101 100 32 116 111 32 106 111 98 32 117 114 103 101 110 116]', '2025-09-11 17:36:00 +0300 EAT', NULL, '15', NULL, NULL, '15', '2025-09-04 22:09:32 +0300 EAT', '15', '2025-09-10 14:37:41 +0300 EAT', '0', NULL, NULL, NULL);
INSERT INTO `jobs` (id, delated, requested_department, descriptions, technical_specifications, materials_required, material_used, instrument, location, status, progress, emergency, comment, customer_comment, customer_approve, supervisor_comment, supervisor_approve, attempted, assigned_department, assigned_to, assigned_by, assigned_date, assigned_comment, assigned_report_time, assigned_referred, attempt_by, attempt_date, attempt_comment, created_by, created_date, updated_by, updated_date, printed, printed_date, printed_by, year) VALUES ('20', '0', '2', '[102 100 102 100 102 100 110 102 102 104 102 104 102 110 102 110 102 102 32]', NULL, NULL, NULL, '[112 97 116 105 101 110 116 32 109 111 110 105 116 111 114]', '[114 111 111 109 32 53]', '[112 101 110 100 105 110 103]', '[48 46 48 48]', '1', '[116 97 115 107 32 114 101 113 117 101 115 116 101 100 32 119 97 105 116 32 102 111 114 32 97 115 115 105 103 110 109 101 110 116]', NULL, '0', NULL, '0', '0', '4', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, '15', '2025-09-10 14:33:48 +0300 EAT', NULL, '2025-09-10 14:33:48 +0300 EAT', '0', NULL, NULL, NULL);
INSERT INTO `jobs` (id, delated, requested_department, descriptions, technical_specifications, materials_required, material_used, instrument, location, status, progress, emergency, comment, customer_comment, customer_approve, supervisor_comment, supervisor_approve, attempted, assigned_department, assigned_to, assigned_by, assigned_date, assigned_comment, assigned_report_time, assigned_referred, attempt_by, attempt_date, attempt_comment, created_by, created_date, updated_by, updated_date, printed, printed_date, printed_by, year) VALUES ('21', '0', '2', '[116 116 97 97 32 104 97 122 105 119 97 107 105 32 104 121 102 114 100 100]', '[106 106 106 106 106 104 104 103]', '[106 104 104 106 104 103 104 103 104 103 104]', NULL, '[116 97 97]', '[101 109 100]', '[99 111 109 112 108 101 116 101 100]', '[48 46 48 48]', '1', '[84 97 115 107 32 97 115 115 105 103 110 101 100 44 32 119 97 105 116 32 102 111 114 32 97 116 116 101 109 112 116 105 110 103]', '[]', '0', '[]', '0', '1', '2', '15', '15', NULL, '[80 114 111 99 101 101 100 32 116 111 32 106 111 98 32 32]', '2025-09-12 01:26:00 +0300 EAT', NULL, '15', NULL, '[102 103 102 102 100 100 115 115 100]', '15', '2025-09-12 11:21:34 +0300 EAT', '15', '2025-09-12 11:25:45 +0300 EAT', '0', NULL, NULL, NULL);

-- Table structure for `login_history`
CREATE TABLE `login_history` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `delated` tinyint(1) NOT NULL DEFAULT 0,
  `status` varchar(20) NOT NULL DEFAULT 'active',
  `ip_address` varchar(100) NOT NULL,
  `login_device` varchar(200) NOT NULL,
  `browser_name` varchar(100) NOT NULL,
  `browser_version` varchar(100) NOT NULL,
  `host` varchar(200) NOT NULL,
  `time_zone` varchar(100) NOT NULL,
  `created_by` int(11) DEFAULT NULL,
  `created_date` datetime DEFAULT current_timestamp(),
  `updated_by` int(11) DEFAULT NULL,
  `updated_date` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `created_by` (`created_by`),
  KEY `updated_by` (`updated_by`),
  CONSTRAINT `login_history_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `login_history_ibfk_2` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `login_history_ibfk_3` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=134 DEFAULT CHARSET=utf8mb4;

INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('58', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-01 04:58:43 +0300 EAT', NULL, '2025-09-01 04:58:43 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('60', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-01 06:35:15 +0300 EAT', NULL, '2025-09-01 06:35:15 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('62', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-01 06:40:32 +0300 EAT', NULL, '2025-09-01 06:40:32 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('64', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-01 07:03:07 +0300 EAT', NULL, '2025-09-01 07:03:07 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('65', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-02 01:54:21 +0300 EAT', NULL, '2025-09-02 01:54:21 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('66', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-02 02:05:44 +0300 EAT', NULL, '2025-09-02 02:05:44 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('67', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-02 03:41:22 +0300 EAT', NULL, '2025-09-02 03:41:22 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('68', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-03 00:54:27 +0300 EAT', NULL, '2025-09-03 00:54:27 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('69', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-03 00:55:12 +0300 EAT', NULL, '2025-09-03 00:55:12 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('70', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 20:52:48 +0300 EAT', NULL, '2025-09-04 20:52:48 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('79', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 21:53:15 +0300 EAT', NULL, '2025-09-04 21:53:15 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('80', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 21:53:46 +0300 EAT', NULL, '2025-09-04 21:53:46 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('81', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 21:54:45 +0300 EAT', NULL, '2025-09-04 21:54:45 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('82', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 21:55:11 +0300 EAT', NULL, '2025-09-04 21:55:11 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('83', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 21:55:57 +0300 EAT', NULL, '2025-09-04 21:55:57 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('84', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-04 22:05:39 +0300 EAT', NULL, '2025-09-04 22:05:39 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('85', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-06 09:08:05 +0300 EAT', NULL, '2025-09-06 09:08:05 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('86', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-06 10:00:30 +0300 EAT', NULL, '2025-09-06 10:00:30 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('87', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-06 10:01:38 +0300 EAT', NULL, '2025-09-06 10:01:38 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('90', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-06 17:53:51 +0300 EAT', NULL, '2025-09-06 17:53:51 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('91', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[49 50 55 46 48 46 48 46 49]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-06 18:03:53 +0300 EAT', NULL, '2025-09-06 18:03:53 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('92', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-06 23:43:29 +0300 EAT', NULL, '2025-09-06 23:43:29 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('93', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-07 03:18:44 +0300 EAT', NULL, '2025-09-07 03:18:44 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('94', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-07 03:19:18 +0300 EAT', NULL, '2025-09-07 03:19:18 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('95', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-07 03:22:16 +0300 EAT', NULL, '2025-09-07 03:22:16 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('96', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-07 03:36:16 +0300 EAT', NULL, '2025-09-07 03:36:16 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('97', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 20:23:35 +0300 EAT', NULL, '2025-09-08 20:23:35 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('98', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 20:44:27 +0300 EAT', NULL, '2025-09-08 20:44:27 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('99', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 20:46:32 +0300 EAT', NULL, '2025-09-08 20:46:32 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('100', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:09:39 +0300 EAT', NULL, '2025-09-08 21:09:39 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('101', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:09:39 +0300 EAT', NULL, '2025-09-08 21:09:39 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('102', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:13:54 +0300 EAT', NULL, '2025-09-08 21:13:54 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('103', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:40:41 +0300 EAT', NULL, '2025-09-08 21:40:41 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('104', '15', '0', '[97 99 116 105 118 101]', '[]', '[]', '[]', '[]', '[]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:41:57 +0300 EAT', NULL, '2025-09-08 21:41:57 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('105', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:42:41 +0300 EAT', NULL, '2025-09-08 21:42:41 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('106', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:47:47 +0300 EAT', NULL, '2025-09-08 21:47:47 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('107', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:55:05 +0300 EAT', NULL, '2025-09-08 21:55:05 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('108', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 21:55:18 +0300 EAT', NULL, '2025-09-08 21:55:18 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('109', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 22:03:20 +0300 EAT', NULL, '2025-09-08 22:03:20 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('110', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 22:03:31 +0300 EAT', NULL, '2025-09-08 22:03:31 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('111', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 22:03:56 +0300 EAT', NULL, '2025-09-08 22:03:56 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('112', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[65 110 100 114 111 105 100 32 54 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-08 22:08:03 +0300 EAT', NULL, '2025-09-08 22:08:03 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('113', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 51 57 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-09 20:52:42 +0300 EAT', NULL, '2025-09-09 20:52:42 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('114', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-09 21:55:29 +0300 EAT', NULL, '2025-09-09 21:55:29 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('115', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 01:16:04 +0300 EAT', NULL, '2025-09-10 01:16:04 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('116', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 10:08:39 +0300 EAT', NULL, '2025-09-10 10:08:39 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('117', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 10:15:48 +0300 EAT', NULL, '2025-09-10 10:15:48 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('118', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 10:20:00 +0300 EAT', NULL, '2025-09-10 10:20:00 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('120', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 10:35:02 +0300 EAT', NULL, '2025-09-10 10:35:02 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('121', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 10:40:08 +0300 EAT', NULL, '2025-09-10 10:40:08 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('122', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 13:28:31 +0300 EAT', NULL, '2025-09-10 13:28:31 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('123', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-10 14:33:04 +0300 EAT', NULL, '2025-09-10 14:33:04 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('124', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-11 03:30:30 +0300 EAT', NULL, '2025-09-11 03:30:30 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('125', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-11 18:15:59 +0300 EAT', NULL, '2025-09-11 18:15:59 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('126', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 10:36:57 +0300 EAT', NULL, '2025-09-12 10:36:57 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('127', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 10:43:02 +0300 EAT', NULL, '2025-09-12 10:43:02 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('128', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 11:09:12 +0300 EAT', NULL, '2025-09-12 11:09:12 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('129', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 11:09:18 +0300 EAT', NULL, '2025-09-12 11:09:18 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('130', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 11:14:57 +0300 EAT', NULL, '2025-09-12 11:14:57 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('131', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 11:26:57 +0300 EAT', NULL, '2025-09-12 11:26:57 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('132', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-12 11:46:41 +0300 EAT', NULL, '2025-09-12 11:46:41 +0300 EAT');
INSERT INTO `login_history` (id, user_id, delated, status, ip_address, login_device, browser_name, browser_version, host, time_zone, created_by, created_date, updated_by, updated_date) VALUES ('133', '15', '0', '[97 99 116 105 118 101]', '[49 50 55 46 48 46 48 46 49]', '[87 105 110 100 111 119 115 32 78 84 32 49 48 46 48]', '[67 104 114 111 109 101]', '[49 52 48 46 48 46 48 46 48]', '[108 111 99 97 108 104 111 115 116]', '[65 102 114 105 99 97 47 68 97 114 95 101 115 95 83 97 108 97 97 109]', '15', '2025-09-13 14:10:18 +0300 EAT', NULL, '2025-09-13 14:10:18 +0300 EAT');

-- Table structure for `maitenances`
CREATE TABLE `maitenances` (
  `id` int(50) NOT NULL AUTO_INCREMENT,
  `created_date` varchar(50) NOT NULL,
  `created_by` varchar(50) NOT NULL,
  `unique_id` varchar(200) NOT NULL,
  `updated_date` varchar(30) NOT NULL,
  `updated_by` varchar(50) NOT NULL,
  `name` varchar(20) NOT NULL,
  `status` varchar(10) NOT NULL,
  `deleted` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- Table structure for `messages`
CREATE TABLE `messages` (
  `id` int(100) NOT NULL AUTO_INCREMENT,
  `sender` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  `receiver` varchar(250) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  `text` varchar(1000) NOT NULL,
  `created_date` varchar(30) NOT NULL,
  `respond_id` varchar(100) NOT NULL,
  `created_by` varchar(100) NOT NULL,
  `updated_date` varchar(30) NOT NULL,
  `updated_by` varchar(100) NOT NULL,
  `status` varchar(30) NOT NULL,
  `description` varchar(100) NOT NULL,
  `sender_id` varchar(100) NOT NULL,
  `delivery` varchar(5) NOT NULL,
  `message_id` varchar(100) NOT NULL,
  `delated` bit(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4;

INSERT INTO `messages` (id, sender, receiver, text, created_date, respond_id, created_by, updated_date, updated_by, status, description, sender_id, delivery, message_id, delated) VALUES ('4', '[123 34 117 115 101 114 95 110 97 109 101 34 58 34 118 97 114 105 97 98 108 101 57 56 64 34 125]', '[123 34 117 115 101 114 95 110 97 109 101 34 58 34 86 97 114 105 97 98 108 101 57 56 57 64 34 44 34 112 104 111 110 101 95 110 117 109 98 101 114 34 58 34 98 34 125]', '[]', '[50 48 50 52 45 48 56 45 48 56 84 49 49 58 53 56 58 48 52 32 53 56 52 32]', '[]', '[118 97 114 105 97 98 108 101 57 56 64]', '[]', '[]', '[]', '[]', '[]', '[]', '[]', '[1]');

-- Table structure for `roles`
CREATE TABLE `roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(200) NOT NULL,
  `descriptions` text DEFAULT NULL,
  `data` longtext NOT NULL,
  `comment` text DEFAULT NULL,
  `status` text NOT NULL DEFAULT 'active',
  `created_by` int(11) DEFAULT NULL,
  `created_date` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_by` int(11) DEFAULT NULL,
  `updated_date` datetime DEFAULT NULL ON UPDATE current_timestamp(),
  `delated` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`),
  KEY `created_by` (`created_by`),
  KEY `updated_by` (`updated_by`),
  CONSTRAINT `roles_ibfk_1` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `roles_ibfk_2` FOREIGN KEY (`updated_by`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COMMENT='Table for managing user roles in the organization';

INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('1', '[117 115 101 114 115]', '[102 111 114 32 115 121 115 116 101 109 32 117 115 101 114 115]', '[50 48 48 48 44 50 48 48 53 44 57 48 48 48 44 49 48 48 52 44 52 48 48 49 44 52 48 48 50 44 52 48 48 51 44 52 48 48 52 44 52 48 48 53 44 52 48 48 54 44 52 48 48 55 44 53 48 48 48 44 53 48 48 49 44 53 48 48 50 44 52 48 48 48 44 53 48 48 52 44 53 48 48 51 44 53 48 48 53 44 50 48 48 55 44 49 48 48 48]', '[]', '[97 99 116 105 118 101]', '15', '2025-05-14 01:04:32 +0300 EAT', '15', '2025-05-18 12:16:27 +0300 EAT', '0');
INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('2', '[69 68]', '[101 120 101 99 117 116 105 118 101 32 100 105 114 101 99 116 111 114 32 114 111 108 101]', '[51 48 48 48 44 51 48 48 49 44 51 48 48 50 44 51 48 48 51 44 51 48 48 53 44 50 48 48 48 44 50 48 48 51 44 51 48 48 52 44 50 48 48 53 44 49 48 48 48 44 49 48 48 52 44 49 48 48 51 44 49 48 48 49 44 49 48 48 50 44 50 48 48 50 44 50 48 48 49 44 50 48 48 55 44 50 48 48 56 44 50 48 48 54 44 50 48 48 52 44 52 48 48 54 44 52 48 48 53 44 53 48 48 48 44 53 48 48 49 44 53 48 48 50 44 53 48 48 53 44 53 48 48 52 44 53 48 48 51 44 50 48 48 57 44 52 48 48 48 44 52 48 48 49 44 52 48 48 55 44 52 48 48 57 44 52 48 48 56 44 52 48 48 50 44 52 48 48 51 44 57 48 48 48 44 50 48 49 48 44 50 48 50 48]', '[]', '[97 99 116 105 118 101]', '15', '2025-05-14 03:35:42 +0300 EAT', '15', '2025-09-02 02:01:40 +0300 EAT', '0');
INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('3', '[72 79 68]', '[72 101 97 100 32 111 102 32 100 101 112 97 114 116 109 101 110 116]', '[50 48 48 48 44 50 48 48 51 44 50 48 48 53 44 57 48 48 48 44 49 48 48 48 44 49 48 48 52 44 50 48 48 54 44 50 48 48 55 44 50 48 48 49 44 50 48 48 50 44 50 48 48 56 44 52 48 48 48 44 52 48 48 54 44 52 48 48 53 44 52 48 48 52 44 52 48 48 50 44 52 48 48 49 44 52 48 48 55 44 53 48 48 48 44 53 48 48 49 44 53 48 48 50 44 53 48 48 53 44 53 48 48 52 44 53 48 48 51 44 52 48 49 48 44 52 48 49 49 44 49 48 48 49 44 49 48 48 50 44 49 48 48 51 44 52 48 48 56 44 52 48 48 57 44 50 48 48 57 44 51 48 48 48 44 52 48 48 51]', '[]', '[97 99 116 105 118 101]', '15', '2025-05-15 08:27:01 +0300 EAT', '15', '2025-08-19 03:15:04 +0300 EAT', '0');
INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('4', '[84 101 99 104 110 105 99 105 97 110]', '[101 110 103 105 110 101 101 114 105 110 103 32 116 101 99 104 110 105 99 105 97 110 32 114 111 108 101]', '[57 48 48 48 44 49 48 48 48 44 49 48 48 52 44 50 48 48 56 44 50 48 48 53 44 52 48 48 48 44 52 48 48 49 44 52 48 48 50 44 52 48 48 51 44 52 48 48 52 44 52 48 48 53 44 52 48 48 54 44 53 48 48 50 44 53 48 48 49 44 53 48 48 48 44 52 48 48 55 44 53 48 48 53 44 53 48 48 52 44 53 48 48 51 44 50 48 48 48 44 50 48 48 55 44 50 48 48 54 44 50 48 48 49 44 50 48 48 50]', '[]', '[97 99 116 105 118 101]', NULL, '2025-05-15 09:54:31 +0300 EAT', NULL, '2025-05-17 09:23:21 +0300 EAT', '0');
INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('5', '[69 110 103 105 110 101 101 114]', '[114 111 108 101 32 102 111 114 32 101 110 103 105 110 101 101 114 32 97 115 115 105 115 116 105 110 103 32 104 101 97 100 32 111 102 32 100 105 114 101 99 116 111 114]', '[52 48 48 52 44 52 48 48 54 44 52 48 48 57 44 52 48 49 52 44 57 48 48 48 44 50 48 48 51 44 50 48 48 53 44 50 48 48 48 44 52 48 48 48 44 52 48 48 49 44 52 48 48 50 44 52 48 48 55 44 52 48 48 53]', '[]', '[97 99 116 105 118 101]', '15', '2025-05-17 09:02:12 +0300 EAT', '15', '2025-08-19 10:45:55 +0300 EAT', '0');
INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('6', '[84 101 115 116]', '[84 101 115 116]', '[57 48 48 48 44 50 48 48 53 44 51 48 48 48 44 51 48 48 49 44 51 48 48 50 44 51 48 48 51 44 51 48 48 52 44 51 48 48 53 44 52 48 48 48 44 52 48 48 49 44 52 48 48 56 44 52 48 48 51 44 52 48 48 50 44 52 48 48 55 44 52 48 48 54 44 52 48 48 57 44 52 48 48 53 44 49 48 48 48 44 49 48 48 49 44 50 48 48 52 44 50 48 48 51 44 50 48 48 50 44 50 48 48 49 44 50 48 48 56 44 50 48 48 57 44 49 48 48 50 44 49 48 48 51 44 49 48 48 52 44 50 48 48 48 44 50 48 48 55 44 50 48 48 54 44 53 48 48 48 44 53 48 48 49 44 53 48 48 53 44 53 48 48 52 44 53 48 48 51 44 53 48 48 50]', '[]', '[97 99 116 105 118 101]', '15', '2025-05-22 01:23:55 +0300 EAT', '15', '2025-08-15 15:23:26 +0300 EAT', '0');
INSERT INTO `roles` (id, name, descriptions, data, comment, status, created_by, created_date, updated_by, updated_date, delated) VALUES ('7', '[115 101 114 105 111 110 32 105 116 32 109 97 110 97 103 101 114]', '[102 102 102 102 102 102 102 102 102]', '[57 48 48 48 44 50 48 48 50 44 50 48 48 51 44 50 48 48 52 44 50 48 48 53 44 50 48 48 54 44 50 48 48 55 44 50 48 48 56 44 50 48 48 57 44 50 48 49 48 44 51 48 48 48 44 51 48 48 49 44 51 48 48 50 44 51 48 48 51 44 51 48 48 52 44 51 48 48 53 44 52 48 48 48]', NULL, '[97 99 116 105 118 101]', '15', '2025-09-10 14:47:40 +0300 EAT', NULL, NULL, '0');

-- Table structure for `setings`
CREATE TABLE `setings` (
  `id` int(10) NOT NULL,
  `created_date` datetime NOT NULL,
  `created_by` int(50) NOT NULL,
  `active_year` int(10) NOT NULL,
  `updated_date` datetime NOT NULL,
  `updated_by` int(11) NOT NULL,
  `name` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- Table structure for `users`
CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `full_name` varchar(100) NOT NULL,
  `user_name` varchar(100) NOT NULL,
  `password` varchar(500) NOT NULL,
  `role_id` varchar(50) DEFAULT NULL,
  `email` varchar(100) NOT NULL,
  `phone_number` varchar(29) NOT NULL,
  `status` varchar(29) NOT NULL,
  `delated` tinyint(1) NOT NULL,
  `created_date` datetime NOT NULL DEFAULT current_timestamp(),
  `gender` varchar(100) NOT NULL,
  `created_by` varchar(100) NOT NULL,
  `updated_date` datetime DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `updated_by` varchar(100) NOT NULL,
  `designation` varchar(100) NOT NULL,
  `department_id` varchar(100) NOT NULL,
  `image` varchar(100) NOT NULL,
  `phone_verify` tinyint(1) NOT NULL DEFAULT 0,
  `email_verify` tinyint(1) NOT NULL DEFAULT 0,
  `otp_code` varchar(100) NOT NULL,
  `address` varchar(200) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_name` (`user_name`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=59 DEFAULT CHARSET=utf8mb4;

INSERT INTO `users` (id, full_name, user_name, password, role_id, email, phone_number, status, delated, created_date, gender, created_by, updated_date, updated_by, designation, department_id, image, phone_verify, email_verify, otp_code, address) VALUES ('15', '[69 110 103 46 32 80 97 116 114 105 99 107 32 77 117 110 117 111]', '[84 115 101 109 117 49 53]', '[66 104 53 80 56 74 47 122 86 97 109 112 109 54 53 111 107 76 121 55 69 71 108 85 97 98 76 89 111 100 47 119 43 110 114 114 82 118 56 115 120 111 56 61]', '[50]', '[112 97 116 114 105 99 107 109 117 110 117 111 57 56 64 103 109 97 105 108 46 99 111 109]', '[43 50 53 53 54 50 53 52 52 57 50 57 53]', '[97 99 116 105 118 101]', '0', '2025-08-11 13:09:30 +0300 EAT', '[109 97 108 101]', '[115 121 115 116 101 109]', '2025-08-27 05:59:11 +0300 EAT', '[110 117 108 108]', '[69 108 101 99 116 114 105 99 97 108 32 69 110 103 105 110 101 101 114]', '[50]', '[110 117 108 108]', '0', '0', '[48]', '[80 46 79 46 66 111 120 32 49 53 55 55 32 77 111 115 104 105]');
INSERT INTO `users` (id, full_name, user_name, password, role_id, email, phone_number, status, delated, created_date, gender, created_by, updated_date, updated_by, designation, department_id, image, phone_verify, email_verify, otp_code, address) VALUES ('34', '[65 122 105 122 105 32 71 101 110 100 111]', '[97 122 105 122 105 46 103 101 110 100 111 64 98 109 104 46 111 114 46 116 122]', '[66 104 53 80 56 74 47 122 86 97 109 112 109 54 53 111 107 76 121 55 69 77 118 47 73 68 106 57 50 110 87 116 72 116 116 88 56 71 74 116 68 51 65 61]', '[49]', '[97 122 105 122 105 46 103 101 110 100 111 64 98 109 104 46 111 114 46 116 122]', '[43 50 53 53 54 50 53 52 52 57 50 57 53]', '[97 99 116 105 118 101]', '0', '2025-08-11 13:09:30 +0300 EAT', '[109 97 108 101]', '[115 121 115 116 101 109]', '2025-09-06 09:57:24 +0300 EAT', '[110 117 108 108]', '[67 105 118 105 108 32 69 110 103 105 110 101 101 114]', '[54 56]', '[110 117 108 108]', '0', '0', '[48 48 48 48]', '[80 46 79 46 66 111 120 32 49 53 55 55 32 77 111 115 104 105]');
INSERT INTO `users` (id, full_name, user_name, password, role_id, email, phone_number, status, delated, created_date, gender, created_by, updated_date, updated_by, designation, department_id, image, phone_verify, email_verify, otp_code, address) VALUES ('56', '[112 97 116 114 105 99 107 32 109 117 110 117 111]', '[66 109 104 50 48 50 53]', '[105 51 81 48 71 65 122 74 66 112 56 48 82 90 68 114 71 51 118 57 72 51 113 111 110 105 89 82 70 119 80 109 109 79 56 112 85 104 75 84 54 107 107 61]', '[50]', '[98 109 104 64 109 97 105 108 46 99 111 109]', '[43 50 53 53 55 54 48 52 52 57 50 57 53]', '[97 99 116 105 118 101]', '0', '2025-09-10 10:33:32 +0300 EAT', '[109 97 108 101]', '[]', '2025-09-11 01:24:59 +0300 EAT', '[]', '[69 108 101 99 116 114 105 99 97 108 32 69 110 103 105 110 101 101 114]', '[52]', '[]', '0', '0', '[]', '[68 111 100 111 109 97 32 109 97 107 117 108 117]');

