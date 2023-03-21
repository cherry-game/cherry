/*
SQLyog Ultimate v12.4.1 (64 bit)
MySQL - 5.7.35 : Database - dev_center
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
CREATE DATABASE /*!32312 IF NOT EXISTS*/`dev_center` /*!40100 DEFAULT CHARACTER SET latin1 */;

USE `dev_center`;

/*Table structure for table `user_bind` */

DROP TABLE IF EXISTS `user_bind`;

CREATE TABLE `user_bind` (
  `uid` bigint(20) NOT NULL COMMENT '用户唯一id',
  `sdk_id` int(11) DEFAULT '0' COMMENT 'sdk配置id',
  `pid` int(11) DEFAULT NULL COMMENT '平台id',
  `open_id` varchar(64) DEFAULT NULL COMMENT '平台帐号open_id',
  `bind_time` bigint(20) DEFAULT NULL COMMENT '绑定时间',
  `up_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后一次更新时间',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `pid_open_id_key` (`pid`,`open_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

/*Data for the table `user_bind` */

insert  into `user_bind`(`uid`,`sdk_id`,`pid`,`open_id`,`bind_time`,`up_time`) values 
(1,1,1,'1',NULL,'2023-03-21 15:17:07'),
(2,2,2,'2',NULL,'2023-03-21 15:17:12');

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
